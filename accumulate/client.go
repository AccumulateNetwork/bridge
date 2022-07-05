package accumulate

import (
	"context"
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

type QueryADIResponse struct {
	Data *ADI `json:"data"`
}

type QueryTokenResponse struct {
	Data *Token `json:"data"`
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
