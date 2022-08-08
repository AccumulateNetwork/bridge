package accumulate

import (
	"path/filepath"
	"strconv"
	"unsafe"
)

// Generate bridge token account in format {chainId}-{symbol}
func GenerateTokenAccount(adi string, chainId int64, symbol string) string {
	return filepath.Join(adi, strconv.Itoa(int(chainId))+"-"+symbol)
}

// Generate bridge data account in format {chainId}:{action}
func GenerateDataAccount(adi string, chainId int64, action string) string {
	return filepath.Join(adi, strconv.Itoa(int(chainId))+":"+action)
}

func byte32(s []byte) (a *[32]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}
