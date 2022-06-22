package accumulate

type Tokens struct {
	Items []*Token `json:"items"`
}

type Token struct {
	AccURL  string          `json:"accURL"`
	Enabled bool            `json:"enabled"`
	Wrapped []*WrappedToken `json:"wrapped"`
}

type WrappedToken struct {
	Address string `json:"address"`
	ChainID int64  `json:"chainId"`
}
