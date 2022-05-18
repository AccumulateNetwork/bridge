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
	GNOSIS_API_MAINNET = "https://safe-transaction.gnosis.io/api/"
	GNOSIS_API_RINKEBY = "https://safe-transaction.rinkeby.gnosis.io/api/"
)

type Gnosis struct {
	API string
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

func NewGnosis(conf *config.Config) (*Gnosis, error) {

	g := &Gnosis{}

	switch conf.EVM.ChainId {

	case 1:
		g.API = GNOSIS_API_MAINNET
		return g, nil
	case 4:
		g.API = GNOSIS_API_RINKEBY
		return g, nil

	}

	return nil, fmt.Errorf("received unknown chainId from config: %s", strconv.Itoa(conf.EVM.ChainId))

}

func (g *Gnosis) GetSafe(safeAddress string) (*ResponseSafe, error) {

	body, err := g.makeRequest("v1/safes/"+safeAddress, nil)
	if err != nil {
		return nil, err
	}

	var safe ResponseSafe

	if err = json.Unmarshal(body, &safe); err != nil {
		return nil, err
	}

	return &safe, nil

}

func (g *Gnosis) makeRequest(path string, req []byte) ([]byte, error) {

	resp, err := http.Post(g.API+path, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
