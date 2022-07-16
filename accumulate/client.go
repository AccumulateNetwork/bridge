package accumulate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/labstack/gommon/log"
)

type ADI struct {
	Type    string `json:"type" validate:"required,eq=identity"`
	KeyBook string `json:"keyBook" validate:"required"`
	URL     string `json:"url" validate:"required"`
}

type Token struct {
	Type      string `json:"type" validate:"required,eq=tokenIssuer"`
	KeyBook   string `json:"keyBook" validate:"required"`
	URL       string `json:"url" validate:"required"`
	Symbol    string `json:"symbol" validate:"required"`
	Precision int64  `json:"precision"`
}

type TokenAccount struct {
	Type     string `json:"type" validate:"required,eq=tokenAccount"`
	KeyBook  string `json:"keyBook" validate:"required"`
	URL      string `json:"url" validate:"required"`
	TokenURL string `json:"tokenUrl" validate:"required"`
	Balance  string `json:"balance" validate:"required"`
}

type DataEntry struct {
	EntryHash string `json:"entryHash" validate:"required"`
	Entry     struct {
		Type string   `json:"type" validate:"required"`
		Data []string `json:"data" validate:"gt=0,dive,required,gt=0"`
	}
}

type Params struct {
	URL      string    `json:"url"`
	Count    int64     `json:"count"`
	Expand   bool      `json:"expand"`
	Envelope *Envelope `json:"envelope"`
}

type Envelope struct {
	Signatures  []*Signature   `json:"signatures"`
	Transaction []*Transaction `json:"transaction"`
}

type Signature struct {
	Type            string `json:"type"`
	PublicKey       string `json:"publicKey"`
	Signature       string `json:"signature"`
	Signer          string `json:"signer"`
	SignerVersion   int64  `json:"signerVersion"`
	Timestamp       int64  `json:"timestamp"`
	TransactionHash string `json:"transactionHash"`
}

type Transaction struct {
	Header TransactionHeader `json:"header"`
	Body   json.RawMessage   `json:"body"`
}

type TransactionHeader struct {
	Principal string `json:"principal"`
	Origin    string `json:"origin"`
	Initiator string `json:"initiator"`
}

type TxSendTokens struct {
	Type string            `json:"type" default:"sendTokens"`
	To   []*TxSendTokensTo `json:"to"`
}

type TxSendTokensTo struct {
	URL    string `json:"url"`
	Amount string `json:"amount"`
}

type ExecuteDirectResponse struct {
	Hash    string `json:"hash"`
	Txid    string `json:"txid"`
	Message string `json:"message"`
}

type QueryADIResponse struct {
	Data *ADI `json:"data"`
}

type QueryTokenResponse struct {
	Data *Token `json:"data"`
}

type QueryTokenAccountResponse struct {
	Data *TokenAccount `json:"data"`
}

type QueryDataResponse struct {
	Data *DataEntry `json:"data"`
}

type QueryDataSetResponse struct {
	Items []*DataEntry `json:"items"`
}

// QueryADI gets Token info
func (c *AccumulateClient) QueryADI(token *Params) (*QueryADIResponse, error) {

	adiResp := &QueryADIResponse{}

	resp, err := c.Client.Call(context.Background(), "query", &token)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(adiResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(adiResp)
	if err != nil {
		return nil, err
	}

	return adiResp, nil

}

// QueryToken gets Token info
func (c *AccumulateClient) QueryToken(token *Params) (*QueryTokenResponse, error) {

	tokenResp := &QueryTokenResponse{}

	resp, err := c.Client.Call(context.Background(), "query", &token)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(tokenResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(tokenResp)
	if err != nil {
		return nil, err
	}

	return tokenResp, nil

}

// QueryTokenAccount gets Token Account info
func (c *AccumulateClient) QueryTokenAccount(account *Params) (*QueryTokenAccountResponse, error) {

	accountResp := &QueryTokenAccountResponse{}

	resp, err := c.Client.Call(context.Background(), "query", &account)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(accountResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(accountResp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return accountResp, nil

}

// QueryData gets Token info
func (c *AccumulateClient) QueryLatestDataEntry(dataAccount *Params) (*QueryDataResponse, error) {

	dataResp := &QueryDataResponse{}

	resp, err := c.Client.Call(context.Background(), "query-data", &dataAccount)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(dataResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(dataResp)
	if err != nil {
		return nil, err
	}

	return dataResp, nil

}

// QueryDataSet gets data entries from data account
func (c *AccumulateClient) QueryDataSet(dataAccount *Params) (*QueryDataSetResponse, error) {

	dataEntriesResp := &QueryDataSetResponse{}

	resp, err := c.Client.Call(context.Background(), "query-data-set", &dataAccount)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(dataEntriesResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(dataEntriesResp)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	return dataEntriesResp, nil

}

// Create calls "execute-direct" tx on Accumulate
func (c *AccumulateClient) ExecuteDirect(params *Params) (*ExecuteDirectResponse, error) {

	callResp := &ExecuteDirectResponse{}

	resp, err := c.Client.Call(context.Background(), "execute-direct", &params)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(callResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	return callResp, nil

}
