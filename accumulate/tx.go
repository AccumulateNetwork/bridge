package accumulate

import (
	"crypto/ed25519"
	"encoding/json"
	"strconv"
	"time"
)

func (c *AccumulateClient) SendTokens(to string, amount int64, tokenURL string, chainId int64) (string, error) {

	// query token
	token, err := c.QueryToken(&Params{URL: tokenURL})
	if err != nil {
		return "", err
	}

	// generate bridge token account for this token
	tokenAccount := GenerateTokenAccount(c.ADI, chainId, token.Data.Symbol)

	// tx
	payload := &TxSendTokens{}
	payload.To = append(payload.To, &TxSendTokensTo{URL: token.Data.URL, Amount: amount})

	tx := &Transaction{}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	tx.Body = body
	tx.Header.Principal = tokenAccount
	tx.Header.Origin = tokenAccount
	tx.Header.Initiator = ""

	// signature
	sig := &Signature{}
	sig.Type = SIGNATURE_TYPE
	sig.PublicKey = string(c.PublicKey)
	sig.Signer = c.ADI + ACC_KEYBOOK + strconv.Itoa(1)
	sig.SignerVersion = 1
	sig.Timestamp = int64(nonceFromTimeNow())
	sig.TransactionHash = "123456"

	sigBytes := ed25519.Sign(c.PrivateKey, []byte(sig.TransactionHash))
	sig.Signature = string(sigBytes)

	e := &Envelope{}
	e.Transaction = append(e.Transaction, tx)
	e.Signatures = append(e.Signatures, sig)

	resp, err := c.ExecuteDirect(e)
	if err != nil {
		return "", err
	}

	return resp.Txid, nil

}

func nonceFromTimeNow() uint64 {
	t := time.Now()
	return uint64(t.Unix()*1e6) + uint64(t.Nanosecond())/1e3
}
