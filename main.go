package main

import (
	"bytes"
	"encoding/hex"
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

	"github.com/labstack/gommon/log"
)

var isLeader bool

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

		fmt.Printf("Accumulate public key hash: %x\n", a.PublicKeyHash)
		fmt.Println("Accumulate API:", a.API)

		die := make(chan bool)
		leaderDataAccount := conf.ACME.BridgeADI + accumulate.ACC_LEADER
		go getLeader(a, leaderDataAccount, die)

		fmt.Println("Getting Accumulate tokens...")
		for _, item := range conf.ACME.Tokens {
			fmt.Println("Trying to get:", item)
			token := &accumulate.QueryTokenResponse{}
			token, err = a.QueryToken(&accumulate.URLRequest{URL: item})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(token.Data)
		}

		// Init Accumulate Bridge API
		fmt.Println("Starting Accumulate Bridge API...")
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.App.APIPort), nil))

	}

}

func getLeader(a *accumulate.AccumulateClient, leaderDataAccount string, die chan bool) {

	var err error

	for {

		select {
		default:

			leaderData := &accumulate.QueryDataResponse{}
			leaderData, err = a.QueryLatestDataEntry(&accumulate.URLRequest{URL: leaderDataAccount})
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
						fmt.Println("[leader] THIS NODE IS LEADER")
					}
					isLeader = true
				}
			}

			// check leader every minute
			time.Sleep(time.Duration(1) * time.Minute)

		case <-die:
			return
		}

	}

}
