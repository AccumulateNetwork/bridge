package gnosis

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

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

/*
type RequestEstSafeTxGas struct {
	To        string `json:"to"`
	Value     int64  `json:"value"`
	Data      string `json:"data"`
	Operation int64  `json:"operation"`
}

type ResponseEstSafeTxGas struct {
	SafeTxGas int64 `json:"safeTxGas,string,omitempty"`
}
*/

type RequestGnosisTx struct {
	Safe                    string  `json:"safe"`
	To                      string  `json:"to"`
	Value                   int64   `json:"value"`
	Data                    string  `json:"data"`
	Operation               int64   `json:"operation"`
	GasToken                string  `json:"gasToken"`
	SafeTxGas               int64   `json:"safeTxGas"`
	BaseGas                 int64   `json:"baseGas"`
	GasPrice                int64   `json:"gasPrice"`
	RefundReceiver          string  `json:"refundReceiver"`
	Nonce                   int64   `json:"nonce"`
	ContractTransactionHash string  `json:"contractTransactionHash"`
	Sender                  string  `json:"sender"`
	Signature               string  `json:"signature"`
	Origin                  *string `json:"origin"`
}

type ResponseErrorGnosisTx struct {
	NonFieldErrors []string `json:"nonFieldErrors"`
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

// CreateSafeMultisigTx submits multisig tx to gnosis safe API
func (g *Gnosis) CreateSafeMultisigTx(data *RequestGnosisTx) (*ResponseErrorGnosisTx, error) {

	param, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	body, err := g.makeRequest("safes/"+g.SafeAddress+"/multisig-transactions/", param)
	if err != nil {
		return nil, err
	}

	var resp ResponseErrorGnosisTx

	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil

}

// internal function that sends API requests
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
