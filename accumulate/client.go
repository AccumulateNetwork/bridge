package accumulate

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/log"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
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

type TokenTx struct {
	From string       `json:"from" validate:"required"`
	To   []*TokenTxTo `json:"to" validate:"required"`
}

type TokenTxTo struct {
	URL    string `json:"url" validate:"required"`
	Amount string `json:"amount" validate:"required"`
}

type DataEntry struct {
	EntryHash string `json:"entryHash" validate:"required"`
	Entry     struct {
		Type string   `json:"type" validate:"required"`
		Data []string `json:"data" validate:"gt=0,dive,required,gt=0"`
	}
}

type Params struct {
	URL      string             `json:"url"`
	Count    int64              `json:"count"`
	Expand   bool               `json:"expand"`
	Envelope *protocol.Envelope `json:"envelope"`
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

type QueryPendingChainResponse struct {
	Items []string `json:"items"`
}

type QueryTokenTxResponse struct {
	Data *TokenTx `json:"data"`
}

type QueryTxHistoryResponse struct {
	Items []*TxHistoryItem `json:"items"`
}

type TxHistoryItem struct {
	TxHash string     `json:"transactionHash"`
	Data   *DataEntry `json:"data"`
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

// QueryTokenTx gets token tx by url
func (c *AccumulateClient) QueryTokenTx(tx *Params) (*QueryTokenTxResponse, error) {

	txResp := &QueryTokenTxResponse{}

	resp, err := c.Client.Call(context.Background(), "query", &tx)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(txResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(txResp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return txResp, nil

}

// QueryTxHistory gets tx history of account
func (c *AccumulateClient) QueryTxHistory(account *Params) (*QueryTxHistoryResponse, error) {

	historyResp := &QueryTxHistoryResponse{}

	resp, err := c.Client.Call(context.Background(), "query", &account)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(historyResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(historyResp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return historyResp, nil

}

// QueryLatestDataEntry gets latest data entry from data account
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

// QueryLatestDataEntry gets latest data entry from data account
func (c *AccumulateClient) QueryDataEntry(dataAccount *Params) (*QueryDataResponse, error) {

	dataResp := &QueryDataResponse{}

	resp, err := c.Client.Call(context.Background(), "query", &dataAccount)
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

	err = c.Validate.StructExcept(dataResp, "Data.EntryHash")
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

// QueryPendingChain gets data pending data from data or token account
func (c *AccumulateClient) QueryPendingChain(account *Params) (*QueryPendingChainResponse, error) {

	pendingResp := &QueryPendingChainResponse{}
	account.URL = GeneratePendingChain(account.URL)

	resp, err := c.Client.Call(context.Background(), "query", &account)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	err = resp.GetObject(pendingResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	err = c.Validate.Struct(pendingResp)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	return pendingResp, nil

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

	// debug
	// p, _ := json.Marshal(resp)
	// fmt.Println(string(p))

	err = resp.GetObject(callResp)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal api response: %s", err)
	}

	return callResp, nil

}
