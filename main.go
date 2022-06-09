package main

import (
	"flag"
	"net/http"
	"os/user"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"

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

		// Init Accumulate Bridge API
		log.Info("Starting Accumulate Bridge API")
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.App.APIPort), nil))

	}

}
