package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/user"
	"strconv"

	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/evm"
	"github.com/AccumulateNetwork/bridge/gnosis"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/labstack/gommon/log"
)

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

		fmt.Println("Using config: ", configFile)

		if conf, err = config.NewConfig(configFile); err != nil {
			log.Fatal(err)
		}

		// init gnosis client
		if g, err = gnosis.NewGnosis(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Gnosis safe: ", g.SafeAddress)
		fmt.Println("Bridge address: ", g.BridgeAddress)
		fmt.Println("Gnosis API: ", g.API)

		// init evm client
		if e, err = evm.NewEVM(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("EVM address: ", e.PublicKey)
		fmt.Println("EVM API: ", e.API)

		// init gnosis client
		if a, err = accumulate.NewAccumulateClient(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Accumulate address: ", hexutil.Encode(a.PublicKey))
		fmt.Println("Accumulate API: ", a.API)

		fmt.Println("Getting Accumulate tokens...")
		for _, item := range conf.ACME.Tokens {
			fmt.Println(item.Symbol)
		}

		// Init Accumulate Bridge API
		fmt.Println("Starting Accumulate Bridge API")
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.App.APIPort), nil))

	}

}
