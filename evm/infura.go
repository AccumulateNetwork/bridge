package evm

import (
	"fmt"

	"github.com/AccumulateNetwork/bridge/config"

	"github.com/ethereum/go-ethereum/ethclient"
)

type InfuraClient struct {
	API    string
	Client ethclient.Client
}

func NewInfuraClient(conf *config.Config) (*InfuraClient, error) {

	if conf.EVM.Node == "" {
		return nil, fmt.Errorf("received empty node from config: %s", conf.EVM.Node)
	}

	client, err := ethclient.Dial(conf.EVM.Node)

	if err != nil {
		return nil, fmt.Errorf("Can't connect to node: %s", conf.EVM.Node)
	} else {
		c := &InfuraClient{}

		c.API = conf.EVM.Node
		c.Client = *client
		return c, nil
	}
}
