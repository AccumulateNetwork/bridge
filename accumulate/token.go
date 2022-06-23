package accumulate

type TokenEntry struct {
	URL     string          `json:"url" validate:"required"`
	Enabled bool            `json:"enabled"`
	Wrapped []*WrappedToken `json:"wrapped" validate:"required"`
}

type WrappedToken struct {
	Address string `json:"address" validate:"required,eth_addr"`
	ChainID int64  `json:"chainId" validate:"required,gt=0"`
}

type TokenList struct {
	Items []*TokenListItem `json:"items"`
}

type TokenListItem struct {
	Token
	TokenEntry
}
