package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os/user"
	"path/filepath"
	"time"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/api"
	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/evm"
	"github.com/AccumulateNetwork/bridge/global"
	"github.com/AccumulateNetwork/bridge/gnosis"
	"github.com/AccumulateNetwork/bridge/schema"
	acmeurl "github.com/AccumulateNetwork/bridge/url"
	"github.com/AccumulateNetwork/bridge/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
)

const LEADER_MIN_DURATION = 2

func main() {

	var err error

	usr, err := user.Current()
	if err != nil {
		log.Error(err)
	}

	configFile := usr.HomeDir + "/.accumulatebridge/config.yaml"

	flag.StringVar(&configFile, "c", configFile, "config.yaml path")
	flag.Parse()

	start(configFile)

}

func start(configFile string) {

	for {

		var err error
		var conf *config.Config
		var g *gnosis.Gnosis
		var e *evm.EVMClient
		var a *accumulate.AccumulateClient

		fmt.Println("Using config:", configFile)

		// init config
		if conf, err = config.NewConfig(configFile); err != nil {
			log.Fatal(err)
		}

		// set log level
		log.SetLevel(log.Lvl(conf.App.LogLevel))

		// init gnosis client
		if g, err = gnosis.NewGnosis(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Gnosis safe:", g.SafeAddress)
		fmt.Println("Bridge address:", g.BridgeAddress)
		fmt.Println("Gnosis API:", g.API)

		// init evm client
		if e, err = evm.NewEVMClient(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("EVM address:", e.PublicKey)
		fmt.Println("EVM API:", e.API)
		fmt.Println("EVM ChainId:", e.ChainId)

		// init gnosis client
		if a, err = accumulate.NewAccumulateClient(conf); err != nil {
			log.Fatal(err)
		}

		// init accumulate client
		fmt.Printf("Accumulate public key: %x\n", a.PublicKey)
		fmt.Printf("Accumulate public key hash: %x\n", a.PublicKeyHash)
		fmt.Println("Accumulate API:", a.API)
		fmt.Println("Bridge ADI:", a.ADI)

		// parse bridge fees
		bridgeFeesDataAccount := filepath.Join(conf.ACME.BridgeADI, accumulate.ACC_BRIDGE_FEES)
		fmt.Println("Getting bridge fees from", bridgeFeesDataAccount)
		fees, err := a.QueryLatestDataEntry(&accumulate.Params{URL: bridgeFeesDataAccount})
		if err != nil {
			fmt.Println("unable to get bridge fees from", bridgeFeesDataAccount)
			log.Fatal(err)
		}

		feesBytes, err := hex.DecodeString(fees.Data.Entry.Data[0])
		if err != nil {
			log.Error("can not decode entry data")
			log.Fatal(err)
		}

		err = json.Unmarshal(feesBytes, &global.BridgeFees)
		if err != nil {
			log.Error("unable to unmarshal entry data")
			log.Fatal(err)
		}

		// set chainId for tokens
		global.Tokens.ChainID = int64(conf.EVM.ChainId)

		fmt.Printf("Mint fee: %.2f%%\n", float64(global.BridgeFees.MintFee)/100)
		fmt.Printf("Burn fee: %.2f%%\n", float64(global.BridgeFees.BurnFee)/100)

		// parse token list from Accumulate
		// only once – when node is started
		// token list is mandatory, so return fatal error in case of error
		tokensDataAccount := filepath.Join(conf.ACME.BridgeADI, accumulate.ACC_TOKEN_REGISTRY)
		fmt.Println("Getting Accumulate tokens from", tokensDataAccount)
		tokens, err := a.QueryDataSet(&accumulate.Params{URL: tokensDataAccount, Count: 1000, Expand: true})
		if err != nil {
			fmt.Println("unable to get token list from", tokensDataAccount)
			log.Fatal(err)
		}

		fmt.Println("Got", len(tokens.Items), "data entry(s)")
		for _, item := range tokens.Items {
			parseToken(a, e, item)
		}

		fmt.Println("Found", len(global.Tokens.Items), "token(s)")

		if len(global.Tokens.Items) == 0 {
			log.Fatal("can not operate without tokens, shutting down")
		}

		// init interval go routines
		die := make(chan bool)
		leaderDataAccount := filepath.Join(conf.ACME.BridgeADI, accumulate.ACC_LEADER)
		go getLeader(a, leaderDataAccount, die)
		// go debugLeader(die)
		go processBurnEvents(a, e, conf.EVM.BridgeAddress, die)

		// init Accumulate Bridge API
		fmt.Println("Starting Accumulate Bridge API at port", conf.App.APIPort)
		log.Fatal(api.StartAPI(conf))

	}

}

// getLeader parses current leader's public key hash from Accumulate data account and compares it with Accumulate key in the config to find out if this node is a leader or not
func getLeader(a *accumulate.AccumulateClient, leaderDataAccount string, die chan bool) {

	for {

		select {
		default:

			leaderData, err := a.QueryLatestDataEntry(&accumulate.Params{URL: leaderDataAccount})
			if err != nil {
				fmt.Println("[leader]", err)
				global.IsLeader = false
				global.LeaderDuration = 0
			} else {
				fmt.Println("[leader] Bridge leader:", leaderData.Data.Entry.Data[0])
				decodedLeader, err := hex.DecodeString(leaderData.Data.Entry.Data[0])
				if err != nil {
					fmt.Println(err)
					global.IsLeader = false
					global.LeaderDuration = 0
				}
				if bytes.Equal(decodedLeader, a.PublicKeyHash) {
					global.LeaderDuration++
					if !global.IsLeader {
						if global.LeaderDuration <= LEADER_MIN_DURATION {
							fmt.Println("[leader] This node is leader, confirmations:", global.LeaderDuration, "of", LEADER_MIN_DURATION)
						}
						if global.LeaderDuration >= LEADER_MIN_DURATION {
							global.IsLeader = true
						}
					}
				} else {
					global.IsLeader = false
					global.LeaderDuration = 0
				}
			}

			// check leader every minute
			time.Sleep(time.Duration(1) * time.Minute)

		case <-die:
			return
		}

	}

}

// parseToken parses data entry with token information received from data account
func parseToken(a *accumulate.AccumulateClient, e *evm.EVMClient, entry *accumulate.DataEntry) {

	fmt.Println("Parsing", entry.EntryHash)

	tokenEntry := &schema.TokenEntry{}

	// check version
	if len(entry.Entry.Data) < 2 {
		log.Debug("looking for at least 2 data fields in entry, found ", len(entry.Entry.Data))
		return
	}

	version, err := hex.DecodeString(entry.Entry.Data[0])
	if err != nil {
		log.Debug("can not decode entry data")
		return
	}

	if !bytes.Equal(version, []byte(accumulate.TOKEN_REGISTRY_VERSION)) {
		log.Debug("entry version is not ", accumulate.TOKEN_REGISTRY_VERSION)
		return
	}

	// convert entry data to bytes
	tokenData, err := hex.DecodeString(entry.Entry.Data[1])
	if err != nil {
		log.Debug("can not decode entry data")
		return
	}

	// try to unmarshal the entry
	err = json.Unmarshal(tokenData, tokenEntry)
	if err != nil {
		log.Debug("unable to unmarshal entry data")
		return
	}

	// if entry is disabled, skip
	if !tokenEntry.Enabled {
		log.Debug("token is disabled")
		return
	}

	// validate token
	validate := validator.New()
	err = validate.Struct(tokenEntry)
	if err != nil {
		log.Debug(err)
		return
	}

	token := &schema.Token{}

	for _, wrappedToken := range tokenEntry.Wrapped {
		// search for current chainid
		if wrappedToken.ChainID == global.Tokens.ChainID {
			err = validate.Struct(wrappedToken)
			if err != nil {
				log.Debug(err)
				return
			}
			token.EVMAddress = wrappedToken.Address
			token.EVMMintTxCost = wrappedToken.MintTxCost
		}
	}

	// if no token address found, error
	if token.EVMAddress == "" {
		log.Debug("can not find token address for chainid ", global.Tokens.ChainID)
		return
	}

	// parse token info from Accumulate
	t, err := a.QueryToken(&accumulate.Params{URL: tokenEntry.URL})
	if err != nil {
		log.Debug("can not get token from accumulate api ", err)
		return
	}

	token.URL = t.Data.URL
	token.Symbol = t.Data.Symbol
	token.Precision = t.Data.Precision

	// check if bridge has token account on this chain for this token
	tokenAccountUrl := accumulate.GenerateTokenAccount(a.ADI, global.Tokens.ChainID, token.Symbol)
	_, err = a.QueryTokenAccount(&accumulate.Params{URL: tokenAccountUrl})
	if err != nil {
		log.Debug("can not get token account ", tokenAccountUrl, " from accumulate api")
		return
	}

	// parse token info from Ethereum
	evmT, err := e.GetERC20(token.EVMAddress)
	if err != nil {
		log.Debug("can not get token from ethereum api ", err)
	}
	token.EVMSymbol = evmT.Symbol
	token.EVMDecimals = evmT.Decimals

	// check for duplicates, if found override
	exists := utils.SearchAccumulateToken(token.URL)

	// if not found, append new token
	if exists == nil {
		log.Info("added token ", token.URL)
		global.Tokens.Items = append(global.Tokens.Items, token)
		return
	}

	log.Info("duplicate token ", token.URL, ", overwritten")
	*exists = *token

}

// debugLeader helps to debug leader behaviour
func debugLeader(die chan bool) {

	for {

		select {
		default:

			log.Debug("isLeader=", global.IsLeader)
			time.Sleep(time.Duration(5) * time.Second)

		case <-die:
			return
		}

	}

}

// processBurnEvents
func processBurnEvents(a *accumulate.AccumulateClient, e *evm.EVMClient, bridge string, die chan bool) {

	for {

		select {
		default:

			if global.IsLeader {

				start := int64(0)
				// start from latest block+1
				// TO DO: get the latest blockheight from Accumulate

				fmt.Println("[burn] Parsing new EVM events for", bridge)
				logs, err := e.ParseBridgeLogs("Burn", bridge, start)
				if err != nil {
					log.Error(err)
				}

				// logs are sorted by timestamp asc
				for _, log := range logs {

					// create burnEntry
					burnEntry := &schema.BurnEvent{}
					burnEntry.EVMTxID = log.TxID.Hex()
					burnEntry.BlockHeight = int64(log.BlockHeight)
					burnEntry.TokenAddress = log.Token.String()
					burnEntry.Destination = log.Destination
					burnEntry.Amount = log.Amount.Int64()

					// find token
					token := utils.SearchEVMToken(burnEntry.TokenAddress)

					// skip if no token found
					if token == nil {
						continue
					}

					fmt.Println("Sending", burnEntry.Amount, token.Symbol, "to", burnEntry.Destination)

					// generate accumulate token tx
					txhash, err := a.SendTokens(burnEntry.Destination, burnEntry.Amount, token.URL, int64(e.ChainId))
					if err != nil {
						fmt.Println("tx failed:", err)
						continue
					}

					fmt.Println("tx sent:", txhash)

					burnEntry.TxHash = txhash

					burnEntryBytes, err := json.Marshal(burnEntry)
					if err != nil {
						fmt.Println("can not marshal burn entry:", err)
						continue
					}

					var content [][]byte
					content = append(content, []byte(accumulate.RELEASE_QUEUE_VERSION))
					content = append(content, burnEntryBytes)

					entryhash, err := a.WriteData(accumulate.GenerateDataAccount(a.ADI, int64(e.ChainId), accumulate.ACC_RELEASE_QUEUE), content)
					if err != nil {
						fmt.Println("data entry creation failed:", err)
						continue
					}

					fmt.Println("data entry created:", entryhash)

				}

			} else {

				dataAccountUrl := accumulate.GenerateDataAccount(a.ADI, int64(e.ChainId), accumulate.ACC_RELEASE_QUEUE) + "#pending/0:1000"
				dataAccount := &accumulate.Params{URL: dataAccountUrl}

				dataSet, err := a.QueryTxHistory(dataAccount)
				if err != nil {
					fmt.Println("can not get pending data entries:", err)
				}

				for _, entry := range dataSet.Items {

					burnEntry := &schema.BurnEvent{}

					// check version
					if len(entry.Data.Entry.Data) < 2 {
						fmt.Println("looking for at least 2 data fields in entry, found", len(entry.Data.Entry.Data))
						continue
					}

					version, err := hex.DecodeString(entry.Data.Entry.Data[0])
					if err != nil {
						fmt.Println("can not decode entry data")
						continue
					}

					if !bytes.Equal(version, []byte(accumulate.RELEASE_QUEUE_VERSION)) {
						fmt.Println("entry version is not ", accumulate.RELEASE_QUEUE_VERSION)
						continue
					}

					// convert entry data to bytes
					burnEventBytes, err := hex.DecodeString(entry.Data.Entry.Data[1])
					if err != nil {
						fmt.Println("can not decode entry data")
						continue
					}

					// try to unmarshal the entry
					err = json.Unmarshal(burnEventBytes, burnEntry)
					if err != nil {
						fmt.Println("unable to unmarshal entry data")
						continue
					}

					// find token
					token := utils.SearchEVMToken(burnEntry.TokenAddress)

					// skip if no token found
					if token == nil {
						continue
					}

					fmt.Println("New pending tx:", burnEntry.TxHash, "- Sending", burnEntry.Amount, token.Symbol, "to", burnEntry.Destination)
					fmt.Println("Checking corresponding EVM tx:", burnEntry.EVMTxID)

					evmTx, err := e.GetTx(burnEntry.EVMTxID)
					if err != nil {
						fmt.Println(err)
						continue
					}

					// parse accumulate txid
					tx, err := acmeurl.ParseTxID(burnEntry.TxHash)
					if err != nil {
						fmt.Println(err)
						continue
					}

					remoteTxHash := tx.Hash()

					// TO DO: validate input data
					abiutil.UnpackBurnTxInputData(evmTx.Data)

					// sign accumulate tx
					txhash, err := a.RemoteTransaction(hex.EncodeToString(remoteTxHash[:]))
					if err != nil {
						fmt.Println("tx failed:", err)
						continue
					}

					fmt.Println("tx sent:", txhash)

					// sign data entry
					txhash, err = a.RemoteTransaction(entry.TxHash)
					if err != nil {
						fmt.Println("tx failed:", err)
						continue
					}

					fmt.Println("tx sent:", txhash)

				}

			}

			time.Sleep(time.Duration(30) * time.Second)

		case <-die:
			return
		}

	}

}
