package accumulate

import (
	"context"
	"fmt"
)

type URLRequest struct {
	URL string `json:"url"`
}

type QueryTokenResponse struct {
	Data struct {
		URL       string `json:"url"`
		Symbol    string `json:"symbol"`
		Precision int64  `json:"precision"`
	}
}

type QueryDataResponse struct {
	Data struct {
		EntryHash string `json:"entryHash"`
		Entry     struct {
			Type string   `json:"type"`
			Data []string `json:"data"`
		}
	}
}

// QueryToken gets Token info
func (c *AccumulateClient) QueryToken(token *URLRequest) (*QueryTokenResponse, error) {

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
func (c *AccumulateClient) QueryLatestDataEntry(dataAccount *URLRequest) (*QueryDataResponse, error) {

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
