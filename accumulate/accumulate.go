package accumulate

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/ybbus/jsonrpc/v3"
)

type AccumulateClient struct {
	API        string
	KeyBook    string
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Client     jsonrpc.RPCClient
}

// NewAccumulateClient constructs the Accumulate client
func NewAccumulateClient(conf *config.Config) (*AccumulateClient, error) {

	c := &AccumulateClient{}

	if conf.ACME.Node == "" {
		return nil, fmt.Errorf("received empty node from config: %s", conf.ACME.Node)
	}

	c.API = conf.ACME.Node
	c.Client = jsonrpc.NewClient(conf.ACME.Node)

	if conf.ACME.KeyBook == "" {
		return nil, fmt.Errorf("received empty keyBook from config: %s", conf.ACME.KeyBook)
	}

	c.KeyBook = conf.ACME.KeyBook

	if conf.ACME.PrivateKey == "" {
		return nil, fmt.Errorf("received empty privateKey from config: %s", conf.ACME.PrivateKey)
	}

	c, err := c.ImportPrivateKey(conf.ACME.PrivateKey)
	if err != nil {
		return nil, err
	}

	return c, nil

}

// ImportPrivateKey imports private key and generates corresponding public key
func (c *AccumulateClient) ImportPrivateKey(pk string) (*AccumulateClient, error) {

	privateKey, err := hex.DecodeString(pk)
	if err != nil {
		return nil, err
	}

	c.PrivateKey = ed25519.PrivateKey(privateKey)
	publicKey, ok := c.PrivateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ed25519")
	}

	c.PublicKey = publicKey

	return c, nil

}
