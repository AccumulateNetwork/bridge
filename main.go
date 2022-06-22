package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os/user"
	"strconv"
	"time"

	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/evm"
	"github.com/AccumulateNetwork/bridge/gnosis"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
)

var isLeader bool
var tokenList accumulate.TokenList

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

		// init gnosis client
		if g, err = gnosis.NewGnosis(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Gnosis safe:", g.SafeAddress)
		fmt.Println("Bridge address:", g.BridgeAddress)
		fmt.Println("Gnosis API:", g.API)

		// init evm client
		if e, err = evm.NewEVM(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("EVM address:", e.PublicKey)
		fmt.Println("EVM API:", e.API)

		// init gnosis client
		if a, err = accumulate.NewAccumulateClient(conf); err != nil {
			log.Fatal(err)
		}

		// init accumulate client
		fmt.Printf("Accumulate public key hash: %x\n", a.PublicKeyHash)
		fmt.Println("Accumulate API:", a.API)

		// parse token list from Accumulate
		// only once – when node is started
		// token list is mandatory, so return fatal error in case of error
		tokensDataAccount := conf.ACME.BridgeADI + "/" + accumulate.ACC_TOKEN_REGISTRY
		fmt.Println("Getting Accumulate tokens from", tokensDataAccount)
		tokens, err := a.QueryDataSet(&accumulate.Params{URL: tokensDataAccount, Count: 1000, Expand: true})
		if err != nil {
			fmt.Println("unable to get token list from", tokensDataAccount)
			log.Fatal(err)
		}

		fmt.Println("Got", len(tokens.Items), "data entries")
		for _, item := range tokens.Items {
			parseToken(a, item)
		}

		fmt.Println("Found", len(tokenList.Items), "tokens")

		// init interval go routines
		die := make(chan bool)
		leaderDataAccount := conf.ACME.BridgeADI + "/" + accumulate.ACC_LEADER
		go getLeader(a, leaderDataAccount, die) // every minute

		// init Accumulate Bridge API
		fmt.Println("Starting Accumulate Bridge API...")
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.App.APIPort), nil))

	}

}

// getLeader parses current leader's public key hash from Accumulate data account and compares it with Accumulate key in the config to find out if this node is a leader or not
func getLeader(a *accumulate.AccumulateClient, leaderDataAccount string, die chan bool) {

	var err error

	for {

		select {
		default:

			leaderData := &accumulate.QueryDataResponse{}
			leaderData, err = a.QueryLatestDataEntry(&accumulate.Params{URL: leaderDataAccount})
			if err != nil {
				fmt.Println("[leader]", err)
				isLeader = false
			} else {
				fmt.Println("[leader] Bridge leader:", leaderData.Data.Entry.Data[0])
				decodedLeader, err := hex.DecodeString(leaderData.Data.Entry.Data[0])
				if err != nil {
					fmt.Println(err)
					isLeader = false
				}
				if bytes.Equal(decodedLeader, a.PublicKeyHash) {
					if !isLeader {
						// print this message only on the first run or when node becomes leader
						fmt.Println("[leader] THIS NODE IS LEADER")
					}
					isLeader = true
				} else {
					isLeader = false
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
func parseToken(a *accumulate.AccumulateClient, entry *accumulate.DataEntry) {

	fmt.Println("Parsing", entry.EntryHash)

	token := &accumulate.TokenEntry{}

	// check version
	if len(entry.Entry.Data) < 2 {
		fmt.Println("no extIds found in entry")
		return
	}
	if entry.Entry.Data[1] != accumulate.TOKEN_REGISTRY_VERSION {
		fmt.Println("entry version is not", accumulate.TOKEN_REGISTRY_VERSION)
		return
	}

	// convert entry data to bytes
	tokenData, err := hex.DecodeString(entry.Entry.Data[0])
	if err != nil {
		fmt.Println("can not decode entry data")
		return
	}

	// try to unmarshal the entry
	err = json.Unmarshal(tokenData, token)
	if err != nil {
		fmt.Println("unable to unmarshal entry data")
		return
	}

	// validate token
	validate := validator.New()
	err = validate.Struct(token)
	if err != nil {
		fmt.Println(err)
		return
	}

	// parse token info from Accumulate
	t, err := a.QueryToken(&accumulate.Params{URL: token.URL})
	if err != nil {
		fmt.Println("can not get token from accumulate api", err)
		return
	}

	newItem := &accumulate.TokenListItem{*t.Data, *token}

	tokenList.Items = append(tokenList.Items, newItem)

	return

}
