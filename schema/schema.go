package schema

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/AccumulateNetwork/bridge/accumulate"
)

// BridgeFees schema
type BridgeFees struct {
	MintFee int64 `json:"mintFee" validate:"gte=0"`
	BurnFee int64 `json:"burnFee" validate:"gte=0"`
}

// TokenEntry is token registry item schema
type TokenEntry struct {
	URL       string          `json:"url" validate:"required"`
	Enabled   bool            `json:"enabled"`
	Symbol    string          `json:"-"` // do not marshal symbol
	Precision int64           `json:"-"` // do not marshal precision
	Wrapped   []*WrappedToken `json:"wrapped" validate:"required"`
}

type WrappedToken struct {
	Address    string  `json:"address" validate:"required,eth_addr"`
	ChainID    int64   `json:"chainId" validate:"required,gt=0"`
	MintTxCost float64 `json:"mintTxCost" validate:"gte=0"`
}

// Tokens is the list of active tokens, used by API
type Tokens struct {
	ChainID int64    `json:"chainId"`
	Items   []*Token `json:"items"`
}

// Token is an item of Tokens{}
type Token struct {
	URL           string  `json:"url"`
	Symbol        string  `json:"symbol"`
	Precision     int64   `json:"precision"`
	EVMAddress    string  `json:"evmAddress"`
	EVMSymbol     string  `json:"evmSymbol"`
	EVMDecimals   int64   `json:"evmDecimals"`
	EVMMintTxCost float64 `json:"evmMintTxCost" validate:"gte=0"`
}

// BurnEvent is an event of token burns on the EVM side
type BurnEvent struct {
	EVMTxID      string `json:"evmTxID"`
	BlockHeight  int64  `json:"blockHeight"`
	TokenAddress string `json:"tokenAddress"`
	Amount       int64  `json:"amount"`
	Destination  string `json:"destination"`
	TokenURL     string `json:"-"`
	TxHash       string `json:"txHash"`
}

// DepositEvent is an event of token deposit into bridge token account
type DepositEvent struct {
	TxID         string `json:"txid"`
	Source       string `json:"source"`
	TokenURL     string `json:"tokenURL"`
	Amount       int64  `json:"amount"`
	SeqNumber    int64  `json:"seqNumber"`
	Destination  string `json:"destination"`
	TokenAddress string `json:"-"`
	SafeTxHash   string `json:"safeTxHash"`
	SafeTxNonce  int64  `json:"safeTxNonce"`
}

// ParseBurnEvent parses accumulate data entry into burn event and validates it
func ParseBurnEvent(entry *accumulate.DataEntry) (*BurnEvent, error) {

	burnEntry := &BurnEvent{}

	// check version
	if len(entry.Entry.Data) < 2 {
		return nil, fmt.Errorf("looking for at least 2 data fields in entry, found %d", len(entry.Entry.Data))
	}

	version, err := hex.DecodeString(entry.Entry.Data[0])
	if err != nil {
		return nil, fmt.Errorf("can not decode entry data")
	}

	if !bytes.Equal(version, []byte(accumulate.RELEASE_QUEUE_VERSION)) {
		return nil, fmt.Errorf("entry version is not %s", accumulate.RELEASE_QUEUE_VERSION)
	}

	// convert entry data to bytes
	burnEventBytes, err := hex.DecodeString(entry.Entry.Data[1])
	if err != nil {
		return nil, fmt.Errorf("can not decode entry data")
	}

	// try to unmarshal the entry
	err = json.Unmarshal(burnEventBytes, burnEntry)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal entry data")
	}

	return burnEntry, nil

}

// ParseDepositEvent parses accumulate data entry into minut event and validates it
func ParseDepositEvent(entry *accumulate.DataEntry) (*DepositEvent, error) {

	mintEntry := &DepositEvent{}

	// check version
	if len(entry.Entry.Data) < 2 {
		return nil, fmt.Errorf("looking for at least 2 data fields in entry, found %d", len(entry.Entry.Data))
	}

	version, err := hex.DecodeString(entry.Entry.Data[0])
	if err != nil {
		return nil, fmt.Errorf("can not decode entry data")
	}

	if !bytes.Equal(version, []byte(accumulate.MINT_QUEUE_VERSION)) {
		return nil, fmt.Errorf("entry version is not %s", accumulate.MINT_QUEUE_VERSION)
	}

	// convert entry data to bytes
	mintEventBytes, err := hex.DecodeString(entry.Entry.Data[1])
	if err != nil {
		return nil, fmt.Errorf("can not decode entry data")
	}

	// try to unmarshal the entry
	err = json.Unmarshal(mintEventBytes, mintEntry)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal entry data")
	}

	return mintEntry, nil

}
