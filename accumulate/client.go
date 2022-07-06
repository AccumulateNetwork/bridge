package accumulate

import (
	"context"
	"fmt"
	"time"

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
	URL    string `json:"url"`
	Count  int64  `json:"count"`
	Expand bool   `json:"expand"`
}

type CreateParams struct {
	Origin  string `json:"origin"`
	Sponsor string `json:"sponsor"`
	Signer  struct {
		PublicKey string `json:"publicKey"`
		Nonce     int64  `json:"nonce"`
	}
	Signature string `json:"signature"`
	KeyPage   struct {
		Height int64 `json:"height"`
		Index  int64 `json:"index"`
	}
	Payload []byte `json:"payload"`
}

type CreateResponse struct {
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

	log.Info(accountResp)

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

// Create sends tx on Accumulate
func (c *AccumulateClient) Create(method string, tx *CreateParams) (*CreateResponse, error) {

	createResp := &CreateResponse{}

	// fill signer info
	tx.Origin = c.ADI
	tx.Sponsor = c.ADI
	tx.Signer.PublicKey = string(c.PublicKeyHash)
	tx.Signer.Nonce = int64(nonceFromTimeNow())
	tx.Signature = ""
	tx.KeyPage.Height = 1
	tx.KeyPage.Index = 0

	resp, err := c.Client.Call(context.Background(), method, &tx)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(createResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	return createResp, nil

}

func nonceFromTimeNow() uint64 {
	t := time.Now()
	return uint64(t.Unix()*1e6) + uint64(t.Nanosecond())/1e3
}
