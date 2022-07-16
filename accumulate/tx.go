package accumulate

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	payload.To = append(payload.To, &TxSendTokensTo{URL: to, Amount: strconv.FormatInt(amount, 10)})
	payload.Type = "sendTokens"

	tx := &Transaction{}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	tx.Body = body
	tx.Header.Principal = tokenAccount
	tx.Header.Origin = tokenAccount
	tx.Header.Initiator = hex.EncodeToString(c.PublicKeyHash)

	// tx hashes
	txHeader, err := json.Marshal(tx.Header)
	if err != nil {
		return "", err
	}

	txHeaderHash := sha256.Sum256(txHeader)
	txBodyHash := sha256.Sum256(tx.Body)
	txDataHash := sha256.Sum256(append(txHeaderHash[:], txBodyHash[:]...))
	txHash := sha256.Sum256(append(c.PublicKeyHash, txDataHash[:]...))

	// timestamp
	ts := int64(nonceFromTimeNow())

	// signature
	sig := &Signature{}
	sig.Type = SIGNATURE_TYPE
	sig.PublicKey = hex.EncodeToString(c.PublicKey)
	sig.Signer = c.Signer
	sig.SignerVersion = 1
	sig.Timestamp = ts
	sig.TransactionHash = hex.EncodeToString(txHash[:])

	sigBytes := ed25519.Sign(c.PrivateKey, []byte(txHash[:]))
	sig.Signature = hex.EncodeToString(sigBytes)

	e := &Envelope{}
	e.Transaction = append(e.Transaction, tx)
	e.Signatures = append(e.Signatures, sig)

	p := &Params{}
	p.Envelope = e

	print, _ := json.Marshal(p)
	fmt.Printf(string(print))

	resp, err := c.ExecuteDirect(p)
	if err != nil {
		return "", err
	}

	return resp.Txid, nil

}

func nonceFromTimeNow() uint64 {
	t := time.Now()
	return uint64(t.Unix()*1e6) + uint64(t.Nanosecond())/1e3
}
