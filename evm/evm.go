package evm

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	INFURA_API_MAINNET = "https://mainnet.infura.io/v3/"
	INFURA_API_RINKEBY = "https://rinkeby.infura.io/v3/"
)

type EVMClient struct {
	API                 string
	ChainId             int
	PrivateKey          *ecdsa.PrivateKey
	PublicKey           common.Address
	InfuraProjectSecret string
	Client              *ethclient.Client
}

// NewEVMClient constructs the EVM client
func NewEVMClient(conf *config.Config) (*EVMClient, error) {

	c := &EVMClient{}

	c.API = conf.EVM.Node
	c.ChainId = conf.EVM.ChainId

	// if Infura ID/secret is in config, use Infura API
	if conf.EVM.InfuraProjectID != "" && conf.EVM.InfuraProjectSecret != "" {

		c.InfuraProjectSecret = conf.EVM.InfuraProjectSecret

		switch c.ChainId {

		case 1:
			c.API = INFURA_API_MAINNET + conf.EVM.InfuraProjectID
		case 4:
			c.API = INFURA_API_RINKEBY + conf.EVM.InfuraProjectID
		default:
			return nil, fmt.Errorf("received unknown chainId from config: %s", strconv.Itoa(c.ChainId))

		}

	}

	client, err := ethclient.Dial(conf.EVM.Node)
	if err != nil {
		return nil, fmt.Errorf("can not connect to node: %s", conf.EVM.Node)
	}

	c.Client = client

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
