package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EVMClient struct {
	API        string
	ChainId    int
	PrivateKey *ecdsa.PrivateKey
	PublicKey  common.Address
	Client     *ethclient.Client
}

// NewEVMClient constructs the EVM client
func NewEVMClient(conf *config.Config) (*EVMClient, error) {

	c := &EVMClient{}

	if conf.EVM.Node == "" {
		return nil, fmt.Errorf("received empty node from config: %s", conf.EVM.Node)
	}

	c.API = conf.EVM.Node

	client, err := ethclient.Dial(conf.EVM.Node)
	if err != nil {
		return nil, fmt.Errorf("can not connect to node: %s", conf.EVM.Node)
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("can not get chainId from node: %s", conf.EVM.Node)
	}

	if conf.EVM.ChainId != int(chainId.Int64()) {
		return nil, fmt.Errorf("chainId from node is %d, chainId from config is %d", chainId, conf.EVM.ChainId)
	}

	c.Client = client
	c.ChainId = int(chainId.Int64())

	if conf.EVM.PrivateKey == "" {
		return nil, fmt.Errorf("received empty privateKey from config: %s", conf.EVM.PrivateKey)
	}

	c, err = c.ImportPrivateKey(conf.EVM.PrivateKey)
	if err != nil {
		return nil, err
	}

	return c, nil

}

// ImportPrivateKey imports private key and generates corresponding public key
func (c *EVMClient) ImportPrivateKey(pk string) (*EVMClient, error) {

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}

	c.PrivateKey = privateKey

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	c.PublicKey = crypto.PubkeyToAddress(*publicKeyECDSA)

	return c, nil

}
