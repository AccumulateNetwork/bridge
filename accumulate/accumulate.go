package accumulate

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/go-playground/validator/v10"
	"github.com/ybbus/jsonrpc/v3"
)

const (
	ACC_KEYPAGE                 = "1"       // bridge ADI keypage
	ACC_LEADER                  = "leader"  // data account: current leader (pubkeyhash)
	ACC_TOKEN_REGISTRY          = "tokens"  // data account: token registry (accumulate token address, evm token address, evm chainid)
	ACC_BRIDGE_FEES             = "fees"    // data account: bridge fees
	ACC_MINT_QUEUE              = "mint"    // data account: mint queue, {chainid}:mint
	ACC_RELEASE_QUEUE           = "release" // data account: release queue, {chainid}:release
	ACC_BRIDGE_STATUS           = "status"  // data account: status (1 = on, 0 = off)
	TOKEN_REGISTRY_VERSION      = "v1"      // validate token registry data entries
	MINT_QUEUE_VERSION          = "v1"      // validate burn events data entries
	RELEASE_QUEUE_VERSION       = "v1"      // validate deposit list data entries
	SIGNATURE_TYPE              = "ed25519"
	ZERO_HASH                   = "0000000000000000000000000000000000000000000000000000000000000000"
	TX_TYPE_SYNTH_TOKEN_DEPOSIT = "syntheticDepositTokens"
	TX_TYPE_SEND_TOKENS         = "sendTokens"
)

type AccumulateClient struct {
	API           string
	ADI           string
	Signer        string
	PrivateKey    ed25519.PrivateKey
	PublicKey     ed25519.PublicKey
	PublicKeyHash []byte
	Client        jsonrpc.RPCClient
	Validate      *validator.Validate
}

func notOlderThanOneMinute(fl validator.FieldLevel) bool {
	// Get the value of the field
	lastBlockTime, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}

	// Calculate the difference between the current time and LastBlockTime
	diff := time.Since(lastBlockTime)

	// Check if the difference is less than or equal to 1 minute
	return diff <= time.Minute
}

// NewAccumulateClient constructs the Accumulate client
func NewAccumulateClient(conf *config.Config) (*AccumulateClient, error) {

	c := &AccumulateClient{}

	// init validator
	c.Validate = validator.New()

	err := c.Validate.RegisterValidation("notOlderThanOneMinute", notOlderThanOneMinute)
	if err != nil {
		return nil, fmt.Errorf("Error registering validation function: %s", err)
	}

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
	_, err = c.QueryADI(&Params{URL: conf.ACME.BridgeADI})
	if err != nil {
		return nil, err
	}

	c.ADI = conf.ACME.BridgeADI
	c.Signer = filepath.Join(conf.ACME.BridgeADI, conf.ACME.KeyBook, ACC_KEYPAGE)

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
