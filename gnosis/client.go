package gnosis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"
)

const (
	ZERO_ADDR          = "0x0000000000000000000000000000000000000000"
	GNOSIS_API_MAINNET = "https://safe-transaction.gnosis.io/api/v1/"
	GNOSIS_API_RINKEBY = "https://safe-transaction.rinkeby.gnosis.io/api/v1/"
)

type Gnosis struct {
	API         string
	SafeAddress string
	PrivateKey  string
}

type ResponseSafe struct {
	Address         string   `json:"address"`
	Nonce           int64    `json:"nonce"`
	Threshold       int64    `json:"threshold"`
	Owners          []string `json:"owners"`
	MasterCopy      string   `json:"masterCopy"`
	Modules         []string `json:"modules"`
	FallbackHandler string   `json:"fallbackHandler"`
	Guard           string   `json:"guard"`
	Version         string   `json:"version"`
}

type RequestEstSafeTxGas struct {
	To        string `json:"to"`
	Value     int64  `json:"value"`
	Data      string `json:"data"`
	Operation int64  `json:"operation"`
}

type ResponseEstSafeTxGas struct {
	SafeTxGas int64 `json:"safeTxGas,string,omitempty"`
}

// NewGnosis constructs the gnosis safe
func NewGnosis(conf *config.Config) (*Gnosis, error) {

	g := &Gnosis{}

	switch conf.EVM.ChainId {

	case 1:
		g.API = GNOSIS_API_MAINNET
	case 4:
		g.API = GNOSIS_API_RINKEBY
	default:
		return nil, fmt.Errorf("received unknown chainId from config: %s", strconv.Itoa(conf.EVM.ChainId))

	}

	if conf.EVM.SafeAddress == "" {
		return nil, fmt.Errorf("received empty safeAddress from config: %s", conf.EVM.SafeAddress)
	}
	g.SafeAddress = conf.EVM.SafeAddress

	if conf.EVM.PrivateKey == "" {
		return nil, fmt.Errorf("received empty privateKey from config: %s", conf.EVM.PrivateKey)
	}
	g.PrivateKey = conf.EVM.PrivateKey

	return g, nil

}

// GetSafe gets safe info and current nonce
func (g *Gnosis) GetSafe() (*ResponseSafe, error) {

	body, err := g.makeRequest("safes/"+g.SafeAddress, nil)
	if err != nil {
		return nil, err
	}

	var safe ResponseSafe

	if err = json.Unmarshal(body, &safe); err != nil {
		return nil, err
	}

	return &safe, nil

}

// EstimateSafeTxGas estimates safe tx costs
func (g *Gnosis) EstimateSafeTxGas(req *RequestEstSafeTxGas) (*ResponseEstSafeTxGas, error) {

	req.To = g.SafeAddress

	reqBytes, _ := json.Marshal(req)

	body, err := g.makeRequest("safes/"+g.SafeAddress+"/multisig-transactions/estimations/", reqBytes)
	if err != nil {
		return nil, err
	}

	var safe ResponseEstSafeTxGas

	if err = json.Unmarshal(body, &safe); err != nil {
		return nil, err
	}

	return &safe, nil

}

// internal
func (g *Gnosis) makeRequest(path string, req []byte) ([]byte, error) {

	var resp *http.Response
	var err error

	if req != nil {
		resp, err = http.Post(g.API+path, "application/json", bytes.NewBuffer(req))
	} else {
		resp, err = http.Get(g.API + path)
	}
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
