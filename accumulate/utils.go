package accumulate

import "strconv"

// Generate bridge token account in format {chainId}-{symbol}
func GenerateTokenAccount(chainId int64, symbol string) string {

	return strconv.Itoa(int(chainId)) + "-" + symbol

}
