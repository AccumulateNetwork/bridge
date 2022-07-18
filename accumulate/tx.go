package accumulate

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/url"

	"gitlab.com/accumulatenetwork/accumulate/pkg/client/signing"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

func (c *AccumulateClient) SendTokens(to string, amount int64, tokenURL string, chainId int64) (string, error) {

	// query token
	token, err := c.QueryToken(&Params{URL: tokenURL})
	if err != nil {
		return "", err
	}

	// generate bridge token account for this token
	fromTokenAccount := GenerateTokenAccount(c.ADI, chainId, token.Data.Symbol)
	fmt.Println("token acc:", fromTokenAccount)

	// tx body
	payload := new(protocol.SendTokens)

	url, err := url.Parse(to)
	if err != nil {
		return "", err
	}
	pUrl := protocol.AccountUrl(url.Host, url.Path)

	amountBigInt := *big.NewInt(amount)
	payload.AddRecipient(pUrl, &amountBigInt)

	env, err := c.buildEnvelope(fromTokenAccount, payload)
	if err != nil {
		return "", err
	}

	json, _ := json.Marshal(env)
	fmt.Println(string(json))

	return "", nil

}

func (c *AccumulateClient) buildEnvelope(fromTokenAccount string, payload protocol.TransactionBody) (*protocol.Envelope, error) {

	fromUrl, err := url.Parse(fromTokenAccount)
	if err != nil {
		return nil, err
	}

	signerUrl, err := url.Parse(c.Signer)
	if err != nil {
		return nil, err
	}

	keypage := protocol.AccountUrl(signerUrl.Host, signerUrl.Path)
	from := protocol.AccountUrl(fromUrl.Host, fromUrl.Path)

	signer := new(signing.Builder)
	signer.SetPrivateKey(c.PrivateKey)
	signer.SetTimestampToNow()
	signer.SetVersion(3)
	signer.SetType(protocol.SignatureTypeED25519)
	signer.SetUrl(keypage)

	txn := new(protocol.Transaction)
	txn.Body = payload
	txn.Header.Principal = from

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
