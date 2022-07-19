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
	accUrl := protocol.AccountUrl(toUrl.Authority, toUrl.Path)

	amountBigInt := *big.NewInt(amount)
	payload.AddRecipient(accUrl, &amountBigInt)

	env, err := c.buildEnvelope(fromTokenAccount, payload)
	if err != nil {
		return "", err
	}

	params := &Params{Envelope: env}

	resp, err := c.ExecuteDirect(params)
	if err != nil {
		return "", err
	}

	return resp.Hash, nil

}

// RemoteTransaction generates remote tx for `execute-direct` API method
func (c *AccumulateClient) RemoteTransaction(txhash string) (string, error) {

	// tx body
	payload := new(protocol.RemoteTransaction)
	hash, err := hex.DecodeString(txhash)
	if err != nil {
		return "", err
	}
	payload.Hash = *byte32(hash)

	env, err := c.buildEnvelope(c.ADI, payload)
	if err != nil {
		return "", err
	}

	params := &Params{Envelope: env}

	resp, err := c.ExecuteDirect(params)
	if err != nil {
		return "", err
	}

	return resp.Hash, nil

}

// WriteData generates writeData tx for `execute-direct` API method
func (c *AccumulateClient) WriteData(dataAccount string, content [][]byte) (string, error) {

	// tx body
	entry := new(protocol.AccumulateDataEntry)
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

	return resp.Hash, nil

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

	signer := new(signing.Builder)
	signer.SetPrivateKey(c.PrivateKey)
	signer.SetTimestamp(nonceFromTimeNow())
	signer.SetVersion(3)
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
