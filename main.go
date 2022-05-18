package main

import (
	"flag"
	"net/http"
	"os/user"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/gnosis"

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

		if conf, err = config.NewConfig(configFile); err != nil {
			log.Fatal(err)
		}

		g, err := gnosis.NewGnosis(conf)
		if err != nil {
			log.Fatal(err)
		}

		safe, err := g.GetSafe()
		if err != nil {
			log.Fatal(err)
		}

		log.Info(safe.Nonce)

		req := &gnosis.RequestEstSafeTxGas{}
		req.Value = 0
		req.Operation = 0

		gas, err := g.EstimateSafeTxGas(req)
		if err != nil {
			log.Fatal(err)
		}

		log.Info(gas)

		mint, err := g.MintTx(&gas.SafeTxGas, &safe.Nonce)
		if err != nil {
			log.Fatal(err)
		}

		log.Info(mint)

		// Init Accumulate Bridge API
		log.Info("Starting Accumulate Bridge API")
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.App.APIPort), nil))

	}

}
