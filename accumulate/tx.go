package accumulate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

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
	tokenAccount := GenerateTokenAccount(chainId, token.Data.Symbol)
	fmt.Println(tokenAccount)

	// tx body
	payload := &TxSendTokens{}
	payload.To = append(payload.To, &TxSendTokensTo{URL: to, Amount: strconv.FormatInt(amount, 10)})
	payload.Type = "sendTokens"
	payload.Hash = ZERO_HASH

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// debug
	fmt.Println("tx body:", string(body))

	// tx
	tx := &Transaction{}

	tx.Body = body
	tx.Header.Principal = tokenAccount
	tx.Header.Origin = tokenAccount
	tx.Header.Initiator = hex.EncodeToString(c.PublicKeyHash)

	txHeader, err := json.Marshal(tx.Header)
	if err != nil {
		return "", err
	}

	// debug
	fmt.Println("tx header:", string(txHeader))

	// header, body hashes
	txHeaderHash := sha256.Sum256(txHeader)
	fmt.Println("tx header hash:", hex.EncodeToString(txHeaderHash[:]))

	txBodyHash := sha256.Sum256(body)
	fmt.Println("tx body hash:", hex.EncodeToString(txBodyHash[:]))

	// tx hash
	sha := sha256.New()
	sha.Write(txHeaderHash[:])
	sha.Write(txBodyHash[:])

	txHash := sha.Sum(nil)
	fmt.Println("tx hash:", hex.EncodeToString(txHash[:]))
	/*

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
	*/

	return "", nil

}

func (c *AccumulateClient) SendTokens2(toADI string, toTokenAccount string, amount int64, tokenURL string, chainId int64) (string, error) {

	// query token
	token, err := c.QueryToken(&Params{URL: tokenURL})
	if err != nil {
		return "", err
	}

	// generate bridge token account for this token
	fromTokenAccount := GenerateTokenAccount(chainId, token.Data.Symbol)
	fmt.Println("token acc:", fromTokenAccount)

	// tx body
	payload := new(protocol.SendTokens)
	url := protocol.AccountUrl(toADI, toTokenAccount)

	amountBigInt := *big.NewInt(amount)
	payload.AddRecipient(url, &amountBigInt)

	env, err := c.buildEnvelope(fromTokenAccount, payload)
	if err != nil {
		return "", err
	}

	json, _ := json.Marshal(env)
	fmt.Println(string(json))

	return "", nil

}

func (c *AccumulateClient) buildEnvelope(fromTokenAccount string, payload protocol.TransactionBody) (*protocol.Envelope, error) {

	signerKeyPage := protocol.AccountUrl(c.ADI, c.KeyPage)
	from := protocol.AccountUrl(c.ADI, fromTokenAccount)

	signer := new(signing.Builder)
	signer.SetPrivateKey(c.PrivateKey)
	signer.SetTimestampToNow()
	signer.SetVersion(1)
	signer.SetType(protocol.SignatureTypeED25519)
	signer.SetUrl(signerKeyPage)

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

func nonceFromTimeNow() uint64 {
	t := time.Now()
	return uint64(t.Unix()*1e6) + uint64(t.Nanosecond())/1e3
}
