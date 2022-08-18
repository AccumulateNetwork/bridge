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
	MintFee int64 `json:"mintFee"`
	BurnFee int64 `json:"burnFee"`
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
	Address    string `json:"address" validate:"required,eth_addr"`
	ChainID    int64  `json:"chainId" validate:"required,gt=0"`
	MintTxCost int64  `json:"mintTxCost" validate:"gte=0"`
}

// Tokens is the list of active tokens, used by API
type Tokens struct {
	ChainID int64    `json:"chainId"`
	Items   []*Token `json:"items"`
}

// Token is an item of Tokens{}
type Token struct {
	URL           string `json:"url"`
	Symbol        string `json:"symbol"`
	Precision     int64  `json:"precision"`
	EVMAddress    string `json:"evmAddress"`
	EVMSymbol     string `json:"evmSymbol"`
	EVMDecimals   int64  `json:"evmDecimals"`
	EVMMintTxCost int64  `json:"evmMintTxCost"`
}

// BurnEvent is an event of token burns
type BurnEvent struct {
	EVMTxID      string `json:"evmTxID"`
	BlockHeight  int64  `json:"blockHeight"`
	TokenAddress string `json:"tokenAddress"`
	Amount       int64  `json:"amount"`
	Destination  string `json:"destination"`
	TokenURL     string `json:"-"`
	TxHash       string `json:"txHash"`
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
