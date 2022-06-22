package accumulate

type Token struct {
	AccURL        string `json:"accURL"`
	Enabled       bool   `json:"enabled"`
	WrappedTokens []*WrappedToken
}

type WrappedToken struct {
	Address string `json:"address"`
	ChainID int64  `json:"chainId"`
}
