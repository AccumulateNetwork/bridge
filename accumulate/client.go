package accumulate

import (
	"context"
	"fmt"
)

type QueryTokenRequest struct {
	URL string `json:"url"`
}

type QueryTokenResponse struct {
	Data struct {
		URL       string `json:"url"`
		Symbol    string `json:"symbol"`
		Precision int64  `json:"precision"`
	}
}

// QueryToken gets Token info
func (c *AccumulateClient) QueryToken(token *QueryTokenRequest) (*QueryTokenResponse, error) {

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
