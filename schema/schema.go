package schema

// BridgeFees schema
type BridgeFees struct {
	MintFee int64 `json:"mintFee"`
	EVMFee  int64 `json:"evmFee"`
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
	Address string `json:"address" validate:"required,eth_addr"`
	ChainID int64  `json:"chainId" validate:"required,gt=0"`
}

// TokenList is the list of active tokens, used by node
type TokenList struct {
	Items []*TokenEntry `json:"items"`
}
