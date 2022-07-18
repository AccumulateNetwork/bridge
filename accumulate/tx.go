package accumulate

import (
	"fmt"
	"log"
	"math/big"

	accurl "github.com/AccumulateNetwork/bridge/url"
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

	return resp.Txid, nil

}

func (c *AccumulateClient) buildEnvelope(fromTokenAccount string, payload protocol.TransactionBody) (*protocol.Envelope, error) {

	fromUrl, err := accurl.Parse(fromTokenAccount)
	if err != nil {
		return nil, err
	}

	from := protocol.AccountUrl(fromUrl.Authority, fromUrl.Path)

	signerUrl, err := accurl.Parse(c.Signer)
	if err != nil {
		return nil, err
	}

	keypage := protocol.AccountUrl(signerUrl.Authority, signerUrl.Path)

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
