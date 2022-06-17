package evm

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ETHEREUM_API = "https://mainnet.infura.io/v3"
	RINKEBY_API  = "https://rinkeby.infura.io/v3"
)

type EVMClient struct {
	API                 string
	ChainId             int
	PrivateKey          *ecdsa.PrivateKey
	PublicKey           common.Address
	InfuraProjectSecret string
}

// NewEVM constructs the EVM client
func NewEVM(conf *config.Config) (*EVMClient, error) {

	c := &EVMClient{}

	c.ChainId = conf.EVM.ChainId

	switch c.ChainId {

	case 1:
		c.API = ETHEREUM_API + "/" + conf.EVM.InfuraProjectID
	case 4:
		c.API = RINKEBY_API + "/" + conf.EVM.InfuraProjectID
	default:
		return nil, fmt.Errorf("received unknown chainId from config: %s", strconv.Itoa(c.ChainId))

	}

	c.InfuraProjectSecret = conf.EVM.InfuraProjectSecret

	if conf.EVM.PrivateKey == "" {
		return nil, fmt.Errorf("received empty privateKey from config: %s", conf.EVM.PrivateKey)
	}

	c, err := c.ImportPrivateKey(conf.EVM.PrivateKey)
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
