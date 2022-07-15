package accumulate

import "strconv"

// Generate bridge token account in format acc://{adi}/{chainId}-{symbol}
func GenerateTokenAccount(adi string, chainId int64, symbol string) string {

	return adi + "/" + strconv.Itoa(int(chainId)) + "-" + symbol

}
