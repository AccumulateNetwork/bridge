package accumulate

type TokenEntry struct {
	URL     string          `json:"url" validate:"required"`
	Enabled bool            `json:"enabled"`
	Wrapped []*WrappedToken `json:"wrapped" required:"true"`
}

type WrappedToken struct {
	Address string `json:"address" required:"true,eth_addr"`
	ChainID int64  `json:"chainId" required:"true,gt=0"`
}

type TokenList struct {
	Items []*TokenListItem `json:"items"`
}

type TokenListItem struct {
	Token
	TokenEntry
}
