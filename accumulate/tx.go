package accumulate

import (
	"encoding/hex"
	"log"
	"math/big"

	accurl "github.com/AccumulateNetwork/bridge/url"
	"gitlab.com/accumulatenetwork/accumulate/pkg/client/signing"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

// SendTokens generates sendTokens tx for `execute-direct` API method
func (c *AccumulateClient) SendTokens(to string, amount int64, tokenURL string, chainId int64) (string, error) {

	// query token
	token, err := c.QueryToken(&Params{URL: tokenURL})
	if err != nil {
		return "", err
	}

	// generate bridge token account for this token
	fromTokenAccount := GenerateTokenAccount(c.ADI, chainId, token.Data.Symbol)

	// tx body
	payload := new(protocol.SendTokens)

	toUrl, err := accurl.Parse(to)
	if err != nil {
		return "", err
	}

	// generate accumulate internal/url data structure and fill it
	accumulateUrl := protocol.AcmeUrl()
	accumulateUrl.Authority = toUrl.Authority
	accumulateUrl.Path = toUrl.Path

	amountBigInt := *big.NewInt(amount)
	payload.AddRecipient(accumulateUrl, &amountBigInt)

	env, err := c.buildEnvelope(fromTokenAccount, payload)
	if err != nil {
		return "", err
	}

	params := &Params{Envelope: env}

	resp, err := c.ExecuteDirect(params)
	if err != nil {
		return "", err
	}

	return resp.Txid, nil

}

// RemoteTransaction generates remote tx for `execute-direct` API method
func (c *AccumulateClient) RemoteTransaction(from string, txhash string) (string, error) {

	// tx body
	payload := new(protocol.RemoteTransaction)
	hash, err := hex.DecodeString(txhash)
	if err != nil {
		return "", err
	}
	payload.Hash = *byte32(hash)

	env, err := c.buildEnvelope(from, payload)
	if err != nil {
		return "", err
	}

	params := &Params{Envelope: env}

	resp, err := c.ExecuteDirect(params)
	if err != nil {
		return "", err
	}

	return resp.Txid, nil

}

// WriteData generates writeData tx for `execute-direct` API method
func (c *AccumulateClient) WriteData(dataAccount string, content [][]byte) (string, error) {

	// tx body
	entry := new(protocol.DoubleHashDataEntry)
	entry.Data = content

	payload := new(protocol.WriteData)
	payload.Entry = entry

	env, err := c.buildEnvelope(dataAccount, payload)
	if err != nil {
		return "", err
	}

	params := &Params{Envelope: env}

	resp, err := c.ExecuteDirect(params)
	if err != nil {
		return "", err
	}

	return resp.Txid, nil

}

func (c *AccumulateClient) buildEnvelope(from string, payload protocol.TransactionBody) (*protocol.Envelope, error) {

	fromUrl, err := accurl.Parse(from)
	if err != nil {
		return nil, err
	}

	principal := protocol.AccountUrl(fromUrl.Authority, fromUrl.Path)

	signerUrl, err := accurl.Parse(c.Signer)
	if err != nil {
		return nil, err
	}

	keypage := protocol.AccountUrl(signerUrl.Authority, signerUrl.Path)

	kpData, err := c.QueryKeyPage(&Params{URL: c.Signer})
	if err != nil {
		return nil, err
	}

	signer := new(signing.Builder)
	signer.SetPrivateKey(c.PrivateKey)
	signer.SetTimestampToNow()
	signer.SetVersion(kpData.Data.Version)
	signer.SetType(protocol.SignatureTypeED25519)
	signer.SetUrl(keypage)

	txn := new(protocol.Transaction)
	txn.Body = payload
	txn.Header.Principal = principal

	sig, err := signer.Initiate(txn)
	if err != nil {
		log.Println("Error : ", err.Error())
		return nil, err
	}

	envelope := new(protocol.Envelope)
	envelope.Transaction = append(envelope.Transaction, txn)
	envelope.Signatures = append(envelope.Signatures, sig)
	envelope.TxHash = append(envelope.TxHash, txn.GetHash()...)

	return envelope, nil
}
