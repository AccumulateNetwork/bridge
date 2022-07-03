package schema

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
