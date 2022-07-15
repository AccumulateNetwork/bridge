package accumulate

import (
	"encoding/json"
	"strconv"

	"github.com/labstack/gommon/log"
)

func (c *AccumulateClient) SendTokens(to string, amount int64, tokenURL string, chainId int64) (string, error) {

	// query token
	token, err := c.QueryToken(&Params{URL: tokenURL})
	if err != nil {
		return "", err
	}

	// generate bridge token account for this token
	tokenAccount := GenerateTokenAccount(c.ADI, chainId, token.Data.Symbol)

	payload := &TxSendTokens{}
	payload.To = append(payload.To, &TxSendTokensTo{URL: token.Data.URL, Amount: amount})

	tx := &Transaction{}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", nil
	}

	tx.Body = body
	tx.Header.Initiator = c.ADI + ACC_KEYBOOK + strconv.Itoa(1)
	tx.Header.Origin = tokenAccount
	tx.Header.Initiator = string(c.PublicKeyHash)

	log.Info(tx)

	return "", nil

}
