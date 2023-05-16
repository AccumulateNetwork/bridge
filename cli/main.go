package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/evm"
	"github.com/AccumulateNetwork/bridge/gnosis"
	"github.com/AccumulateNetwork/bridge/schema"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/urfave/cli/v2"
	// imports as package "cli"
)

func main() {

	app := &cli.App{
		Name:  "accbridge",
		Usage: "Accumulate Bridge CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "mint",
				Usage: "Generates and signs tx to mint wrapped token",
				Action: func(c *cli.Context) error {

					if c.NArg() != 3 {
						printMintHelp()
						return nil
					}

					token := c.Args().Get(0)
					recipient := c.Args().Get(1)
					amount, err := strconv.ParseInt(c.Args().Get(2), 10, 64)
					if err != nil {
						fmt.Print("incorrect amount: ")
						return err
					}

					var conf *config.Config
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					g, err := gnosis.NewGnosis(conf)
					if err != nil {
						fmt.Print("can not init gnosis module: ")
						return err
					}

					safe, err := g.GetSafe()
					if err != nil {
						fmt.Print("can not get gnosis safe: ")
						return err
					}

					data, err := abiutil.GenerateMintTxData(token, recipient, big.NewInt(amount))
					if err != nil {
						fmt.Print("can not generate mint tx: ")
						return err
					}

					contractHash, signature, err := g.SignMintTx(token, recipient, big.NewInt(amount))
					if err != nil {
						fmt.Print("can not sign mint tx: ")
						return err
					}

					tx := gnosis.NewMultisigTx{}
					tx.To = g.BridgeAddress
					tx.Data = hexutil.Encode(data)
					tx.GasToken = abiutil.ZERO_ADDR
					tx.RefundReceiver = abiutil.ZERO_ADDR
					tx.Nonce = safe.Nonce
					tx.ContractTransactionHash = hexutil.Encode(contractHash)
					tx.Sender = g.PublicKey.Hex()
					tx.Signature = hexutil.Encode(signature)

					err = g.CreateSafeMultisigTx(&tx)
					if err != nil {
						fmt.Print("gnosis safe api error: ")
						return err
					}

					return nil

				},
			},
			{
				Name:  "eth-submit",
				Usage: "Submits ethereum tx from gnosis safe",
				Action: func(c *cli.Context) error {

					var err error

					if c.NArg() == 0 {
						printEthSubmitHelp()
						return nil
					}

					safeTxHash := c.Args().Get(0)

					var conf *config.Config
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					// setup evm client
					cl, err := evm.NewEVMClient(conf)
					if err != nil {
						fmt.Print("can not init evm client: ")
						return err
					}

					gasPrice := cl.MaxGasFee
					priorityFee := cl.MaxPriorityFee

					// parse optional args
					if c.Args().Get(1) != "" {
						gasPrice, err = strconv.ParseFloat(c.Args().Get(1), 64)
						if err != nil {
							fmt.Print("incorrect gas price: ")
							return err
						}
					}
					if c.Args().Get(2) != "" {
						priorityFee, err = strconv.ParseFloat(c.Args().Get(2), 64)
						if err != nil {
							fmt.Print("incorrect priority fee: ")
							return err
						}
					}

					// init gnosis safe client
					g, err := gnosis.NewGnosis(conf)
					if err != nil {
						fmt.Print("can not init gnosis module: ")
						return err
					}

					// get txs from gnosis
					gnosisTx, err := g.GetSafeMultisigTx(safeTxHash)
					if err != nil {
						fmt.Printf("can not get gnosis safe tx with hash %s: ", safeTxHash)
						return err
					}

					// check if tx is executed
					if gnosisTx.IsExecuted {
						fmt.Printf("tx is already executed")
						return fmt.Errorf("tx is already executed")
					}

					sort.Slice(gnosisTx.Confirmations, func(i, j int) bool {
						switch strings.Compare(strings.ToLower(gnosisTx.Confirmations[i].Owner), strings.ToLower(gnosisTx.Confirmations[j].Owner)) {
						case -1:
							return true
						case 1:
							return false
						}
						return strings.ToLower(gnosisTx.Confirmations[i].Owner) > strings.ToLower(gnosisTx.Confirmations[j].Owner)
					})

					// concatenate signatures
					var sig []byte
					for _, con := range gnosisTx.Confirmations {
						sigBytes, err := hexutil.Decode(con.Signature)
						if err != nil {
							fmt.Print("can not decode signature hex: ")
							return err
						}
						sig = append(sig, sigBytes...)
					}

					// generate tx input data
					txData, err := abiutil.GenerateExecTransaction(g.BridgeAddress, gnosisTx.Data, hexutil.Encode(sig))
					if err != nil {
						fmt.Print("can not generate tx data: ")
						return err
					}

					to := common.HexToAddress(g.SafeAddress)

					// submit ethereum tx
					sentTx, err := cl.SubmitEIP1559Tx(gnosis.MINT_GAS_LIMIT, gasPrice, priorityFee, &to, 0, txData)
					if err != nil {
						return err
					}

					fmt.Printf("tx sent: %s", sentTx.Hash().Hex())

					return nil

				},
			},
			{
				Name:  "release",
				Usage: "Generates and submits tx to release native tokens",
				Action: func(c *cli.Context) error {

					if c.NArg() != 3 {
						printReleaseHelp()
						return nil
					}

					token := c.Args().Get(0)
					to := c.Args().Get(1)
					amount, err := strconv.ParseInt(c.Args().Get(2), 10, 64)
					if err != nil {
						fmt.Print("incorrect amount: ")
						return err
					}

					var conf *config.Config
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					txhash, err := a.SendTokens(to, amount, token, int64(conf.EVM.ChainId))
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				Name:  "acc-sign",
				Usage: "Signs existing accumulate tx",
				Action: func(c *cli.Context) error {

					if c.NArg() != 1 {
						printAccSignHelp()
						return nil
					}

					txid := c.Args().Get(0)

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					txhash, err := a.RemoteTransaction(conf.ACME.BridgeADI, txid)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				// CLI only supports single network/chainid generation
				// Need to upgrade this method after the bridge switches to multiple networks
				Name:  "token-register",
				Usage: "Generates and submits accumulate data entry for token register",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "disable"},
				},
				Action: func(c *cli.Context) error {

					if c.NArg() < 2 {
						printTokenRegisterHelp()
						return nil
					}

					accumulateURL := c.Args().Get(0)
					wrapped := c.Args().Slice()

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					token := &schema.TokenEntry{}
					token.URL = accumulateURL
					token.Enabled = true

					// parse --disable flag if exists
					disabled := c.Bool("disable")
					if disabled {
						token.Enabled = false
					}

					for i := 1; i <= len(wrapped)-1; i++ {

						wrappedToken := &schema.WrappedToken{}

						err := json.Unmarshal([]byte(wrapped[i]), &wrappedToken)

						if err != nil {
							fmt.Println(err)
							return err
						}

						if wrappedToken.ChainID == 0 {
							fmt.Print("chainId must be a number")
							return err
						}

						if wrappedToken.MintTxCost == 0 {
							fmt.Print("mintTxCost must be a number")
							return err
						}

						token.Wrapped = append(token.Wrapped, wrappedToken)

					}

					tokenBytes, err := json.Marshal(token)
					if err != nil {
						fmt.Print(err)
					}

					var content [][]byte
					content = append(content, []byte(accumulate.TOKEN_REGISTRY_VERSION))
					content = append(content, tokenBytes)

					dataAccount := filepath.Join(a.ADI, accumulate.ACC_TOKEN_REGISTRY)

					txhash, err := a.WriteData(dataAccount, content)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				Name:  "update-fees",
				Usage: "Generates and submits accumulate data entry for bridge fees",
				Action: func(c *cli.Context) error {

					if c.NArg() != 2 {
						printUpdateFeesHelp()
						return nil
					}

					mintFeeString := c.Args().Get(0)
					burnFeeString := c.Args().Get(1)

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					mintFee, err := strconv.Atoi(mintFeeString)
					if err != nil {
						fmt.Print("mintFee must be a number")
						return err
					}

					burnFee, err := strconv.Atoi(burnFeeString)
					if err != nil {
						fmt.Print("burnFee must be a number")
						return err
					}

					fees := &schema.BridgeFees{}
					fees.BurnFee = int64(burnFee)
					fees.MintFee = int64(mintFee)

					feesBytes, err := json.Marshal(fees)
					if err != nil {
						fmt.Print(err)
					}

					var content [][]byte
					content = append(content, feesBytes)

					dataAccount := filepath.Join(a.ADI, accumulate.ACC_BRIDGE_FEES)

					txhash, err := a.WriteData(dataAccount, content)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				Name:  "set-leader",
				Usage: "Generates and submits accumulate data entry for bridge leader",
				Action: func(c *cli.Context) error {

					if c.NArg() != 1 {
						printSetLeaderHelp()
						return nil
					}

					leader := c.Args().Get(0)

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					leaderBytes, err := hex.DecodeString(leader)
					if err != nil {
						fmt.Print(err)
					}

					var content [][]byte
					content = append(content, leaderBytes)

					dataAccount := filepath.Join(a.ADI, accumulate.ACC_LEADER)

					txhash, err := a.WriteData(dataAccount, content)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				Name:  "set-release-height",
				Usage: "Generates and submits accumulate data entry for release queue",
				Action: func(c *cli.Context) error {

					if c.NArg() != 2 {
						printSetReleaseHeightHelp()
						return nil
					}

					chainIdString := c.Args().Get(0)
					blockHeightString := c.Args().Get(1)

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					chainId, err := strconv.Atoi(chainIdString)
					if err != nil {
						fmt.Print("chainId must be a number")
						return err
					}

					blockHeight, err := strconv.Atoi(blockHeightString)
					if err != nil {
						fmt.Print("blockHeight must be a number")
						return err
					}

					entry := &schema.BurnEvent{}
					entry.BlockHeight = int64(blockHeight)

					entryBytes, err := json.Marshal(entry)
					if err != nil {
						fmt.Print(err)
					}

					var content [][]byte
					content = append(content, []byte(accumulate.RELEASE_QUEUE_VERSION))
					content = append(content, entryBytes)

					dataAccount := accumulate.GenerateReleaseDataAccount(a.ADI, int64(chainId), accumulate.ACC_RELEASE_QUEUE)

					txhash, err := a.WriteData(dataAccount, content)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				Name:  "set-mint-height",
				Usage: "Generates and submits accumulate data entry for mint queue",
				Action: func(c *cli.Context) error {

					if c.NArg() != 3 {
						printSetMintHeightHelp()
						return nil
					}

					chainIdString := c.Args().Get(0)
					symbol := c.Args().Get(1)
					seqNumberString := c.Args().Get(2)

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					chainId, err := strconv.Atoi(chainIdString)
					if err != nil {
						fmt.Print("chainId must be a number")
						return err
					}

					seqNumber, err := strconv.Atoi(seqNumberString)
					if err != nil {
						fmt.Print("seqNumber must be a number")
						return err
					}

					entry := &schema.DepositEvent{}
					entry.SeqNumber = int64(seqNumber)

					entryBytes, err := json.Marshal(entry)
					if err != nil {
						fmt.Print(err)
					}

					var content [][]byte
					content = append(content, []byte(accumulate.MINT_QUEUE_VERSION))
					content = append(content, entryBytes)

					dataAccount := accumulate.GenerateMintDataAccount(a.ADI, int64(chainId), accumulate.ACC_MINT_QUEUE, symbol)

					txhash, err := a.WriteData(dataAccount, content)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				Name:  "pause",
				Usage: "Generates and submits accumulate data entry to pause the bridge",
				Action: func(c *cli.Context) error {

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					var content [][]byte
					var empty []byte
					content = append(content, empty)

					dataAccount := filepath.Join(a.ADI, accumulate.ACC_BRIDGE_STATUS)

					txhash, err := a.WriteData(dataAccount, content)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
			{
				Name:  "start",
				Usage: "Generates and submits accumulate data entry to start the bridge",
				Action: func(c *cli.Context) error {

					var conf *config.Config
					var err error
					configFile := c.String("config")

					if configFile == "" {
						usr, err := user.Current()
						if err != nil {
							return err
						}
						configFile = usr.HomeDir + "/.accumulatebridge/config.yaml"
					}

					fmt.Printf("using config: %s\n", configFile)

					if conf, err = config.NewConfig(configFile); err != nil {
						fmt.Print("can not load config: ")
						return err
					}

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					var content [][]byte
					content = append(content, []byte("1"))

					dataAccount := filepath.Join(a.ADI, accumulate.ACC_BRIDGE_STATUS)

					txhash, err := a.WriteData(dataAccount, content)
					if err != nil {
						fmt.Print("tx failed: ")
						return err
					}

					fmt.Printf("tx sent: %s", txhash)

					return nil

				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func printMintHelp() {
	fmt.Println("mint [token] [recipient] [amount]")
}

func printEthSubmitHelp() {
	fmt.Println("eth-submit [gnosis safetxhash] [max gas fee (optional)] [max priority fee (optional)]")
}

func printReleaseHelp() {
	fmt.Println("release [token] [recipient] [amount]")
}

func printAccSignHelp() {
	fmt.Println("acc-sign [txid]")
}

func printTokenRegisterHelp() {
	fmt.Println("token-register [accumulate token URL] '{\"address\":\"\",\"chainId\":\"\",\"mintTxCost\":\"\"}'... [--disable (optional)]")
}

func printUpdateFeesHelp() {
	fmt.Println("update-fees [mint fee (bps)] [burn fee (bps)]")
}

func printSetReleaseHeightHelp() {
	fmt.Println("set-release-height [evm chain id] [evm blockheight]")
}

func printSetMintHeightHelp() {
	fmt.Println("set-mint-height [evm chain id] [token symbol] [tx history seq number]")
}

func printSetLeaderHelp() {
	fmt.Println("set-leader [public key hash]")
}
