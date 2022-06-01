package gnosis

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ZERO_ADDR          = "0x0000000000000000000000000000000000000000"
	GNOSIS_API_MAINNET = "https://safe-transaction.gnosis.io/api/v1/"
	GNOSIS_API_RINKEBY = "https://safe-transaction.rinkeby.gnosis.io/api/v1/"
)

type Gnosis struct {
	API           string
	ChainId       int
	SafeAddress   string
	BridgeAddress string
	PrivateKey    *ecdsa.PrivateKey
	PublicKey     common.Address
}

// NewGnosis constructs the gnosis safe
func NewGnosis(conf *config.Config) (*Gnosis, error) {

	g := &Gnosis{}

	g.ChainId = conf.EVM.ChainId

	switch g.ChainId {

	case 1:
		g.API = GNOSIS_API_MAINNET
	case 4:
		g.API = GNOSIS_API_RINKEBY
	default:
		return nil, fmt.Errorf("received unknown chainId from config: %s", strconv.Itoa(g.ChainId))

	}

	if conf.EVM.SafeAddress == "" {
		return nil, fmt.Errorf("received empty safeAddress from config: %s", conf.EVM.SafeAddress)
	}
	g.SafeAddress = conf.EVM.SafeAddress

	if conf.EVM.BridgeAddress == "" {
		return nil, fmt.Errorf("received empty bridgeAddress from config: %s", conf.EVM.BridgeAddress)
	}
	g.BridgeAddress = conf.EVM.BridgeAddress

	if conf.EVM.PrivateKey == "" {
		return nil, fmt.Errorf("received empty privateKey from config: %s", conf.EVM.PrivateKey)
	}

	g, err := g.ImportPrivateKey(conf.EVM.PrivateKey)
	if err != nil {
		return nil, err
	}

	return g, nil

}

// ImportPrivateKey imports private key and generates corresponding public key
func (g *Gnosis) ImportPrivateKey(pk string) (*Gnosis, error) {

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}

	g.PrivateKey = privateKey

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	g.PublicKey = crypto.PubkeyToAddress(*publicKeyECDSA)

	return g, nil

}
