package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/user"
	"sort"
	"strconv"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/gnosis"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/urfave/cli/v2" // imports as package "cli"
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

					data, err := abiutil.GenerateMintTx(token, recipient, big.NewInt(amount))
					if err != nil {
						fmt.Print("can not generate mint tx: ")
						return err
					}

					contractHash, signature, err := g.SignMintTx(token, recipient, big.NewInt(amount))
					if err != nil {
						fmt.Print("can not sign mint tx: ")
						return err
					}

					tx := gnosis.RequestGnosisTx{}
					tx.To = g.BridgeAddress
					tx.Data = hexutil.Encode(data)
					tx.GasToken = gnosis.ZERO_ADDR
					tx.RefundReceiver = gnosis.ZERO_ADDR
					tx.Nonce = safe.Nonce
					tx.ContractTransactionHash = hexutil.Encode(contractHash)
					tx.Sender = g.PublicKey.Hex()
					tx.Signature = hexutil.Encode(signature)

					resp, err := g.CreateSafeMultisigTx(&tx)
					if err != nil {
						fmt.Print("gnosis safe api error: ")
						return err
					}

					fmt.Print(resp)

					return nil

				},
			},
			{
				Name:  "eth-submit",
				Usage: "Submits ethereum tx from gnosis safe",
				Action: func(c *cli.Context) error {

					if c.NArg() != 3 {
						printEthSubmitHelp()
						return nil
					}

					nonce, err := strconv.ParseInt(c.Args().Get(0), 10, 64)
					if err != nil {
						fmt.Print("incorrect nonce: ")
						return err
					}

					gasPrice, err := strconv.ParseInt(c.Args().Get(0), 10, 64)
					if err != nil {
						fmt.Print("incorrect gas price: ")
						return err
					}

					priorityFee, err := strconv.ParseInt(c.Args().Get(2), 10, 64)
					if err != nil {
						fmt.Print("incorrect priority fee: ")
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

					fmt.Print(nonce, gasPrice, priorityFee, safe)

					return nil

				},
			},
			{
				Name:  "redeem",
				Usage: "Generates, signs and submits tx to redeem native tokens",
				Action: func(c *cli.Context) error {

					if c.NArg() != 3 {
						printRedeemHelp()
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

					a, err := accumulate.NewAccumulateClient(conf)
					if err != nil {
						fmt.Print("can not init accumulate client: ")
						return err
					}

					fmt.Print(a, token, recipient, amount)

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

					fmt.Print(a, txid)

					return nil

				},
			},
			{
				Name:  "token-register",
				Usage: "Generates accumulate data entry for token register",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "disable"},
				},
				Action: func(c *cli.Context) error {

					if c.NArg() != 3 {
						printTokenRegisterHelp()
						return nil
					}

					accumulateURL := c.Args().Get(0)
					evmChainIdString := c.Args().Get(1)
					evmTokenContract := c.Args().Get(2)

					var err error

					evmChainId, err := strconv.Atoi(evmChainIdString)
					if err != nil {
						fmt.Print("chainId must be a number")
						return err
					}

					token := &accumulate.TokenEntry{}
					token.URL = accumulateURL
					token.Enabled = true

					// parse --disable flag if exists
					disabled := c.Bool("disable")
					if disabled {
						token.Enabled = false
					}

					token.Wrapped = append(token.Wrapped, &accumulate.WrappedToken{Address: evmTokenContract, ChainID: int64(evmChainId)})

					tokenBytes, err := json.Marshal(token)
					if err != nil {
						fmt.Print(err)
					}

					fmt.Println(hex.EncodeToString([]byte(accumulate.TOKEN_REGISTRY_VERSION)))
					fmt.Println(hex.EncodeToString(tokenBytes))

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
	fmt.Println("eth-submit [gnosis safe tx nonce] [max gwei price] [max priority fee]")
}

func printRedeemHelp() {
	fmt.Println("redeem [token] [recipient] [amount]")
}

func printAccSignHelp() {
	fmt.Println("acc-sign [txid]")
}

func printTokenRegisterHelp() {
	fmt.Println("token-register [accumulate token URL] [evm chain id] [evm token contract address] [--disable (optional)]")
}
