package accumulate

import (
	"context"
	"fmt"
)

type Params struct {
	URL    string `json:"url"`
	Count  int64  `json:"count"`
	Expand bool   `json:"expand"`
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

type DataEntry struct {
	EntryHash string `json:"entryHash"`
	Entry     struct {
		Type string   `json:"type"`
		Data []string `json:"data"`
	}
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
	if err != nil || tokenResp.Data.URL == "" || tokenResp.Data.Symbol == "" {
		return nil, fmt.Errorf("can not unmarshal api response")
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
	if err != nil || dataResp.Data.EntryHash == "" || dataResp.Data.Entry.Data[0] == "" {
		fmt.Println(dataResp.Data)
		return nil, fmt.Errorf("can not unmarshal api response")
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
		fmt.Println(err)
		return nil, fmt.Errorf("can not unmarshal api response")
	}

	return dataEntriesResp, nil

}
