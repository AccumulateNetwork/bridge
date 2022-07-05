package accumulate

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/go-playground/validator/v10"
	"github.com/ybbus/jsonrpc/v3"
)

const (
	ACC_KEYBOOK            = "book"   // bridge ADI keybook
	ACC_LEADER             = "leader" // data account: current leader (pubkeyhash)
	ACC_TOKEN_REGISTRY     = "tokens" // data account: token registry (accumulate token address, evm token address, evm chainid)
	ACC_BRIDGE_FEES        = "fees"   // data account: bridge fees
	TOKEN_REGISTRY_VERSION = "v1"     // validate token registry data entries
)

type AccumulateClient struct {
	API           string
	ADI           string
	PrivateKey    ed25519.PrivateKey
	PublicKey     ed25519.PublicKey
	PublicKeyHash []byte
	Client        jsonrpc.RPCClient
	Validate      *validator.Validate
}

// NewAccumulateClient constructs the Accumulate client
func NewAccumulateClient(conf *config.Config) (*AccumulateClient, error) {

	c := &AccumulateClient{}

	// init validator
	c.Validate = validator.New()

	if conf.ACME.Node == "" {
		return nil, fmt.Errorf("received empty node from config: %s", conf.ACME.Node)
	}

	c.API = conf.ACME.Node

	// 5 seconds timeout
	opts := &jsonrpc.RPCClientOpts{}
	opts.HTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	c.Client = jsonrpc.NewClientWithOpts(conf.ACME.Node, opts)

	// check if config ADI is valid
	adi, err := c.QueryADI(&Params{URL: conf.ACME.BridgeADI})
	if err != nil {
		return nil, err
	}

	c.ADI = adi.Data.URL

	if conf.ACME.PrivateKey == "" {
		return nil, fmt.Errorf("received empty privateKey from config: %s", conf.ACME.PrivateKey)
	}

	c, err = c.ImportPrivateKey(conf.ACME.PrivateKey)
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

	publicKeyHash := sha256.Sum256(c.PublicKey)
	c.PublicKeyHash = publicKeyHash[:]

	return c, nil

}
