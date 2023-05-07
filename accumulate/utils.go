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
func GenerateReleaseDataAccount(adi string, chainId int64, action string) string {
	return filepath.Join(adi, "audit:"+strconv.Itoa(int(chainId))+":"+action)
}

// Generate bridge data account in format {chainId}:{action}:{symbol}
func GenerateMintDataAccount(adi string, chainId int64, action string, symbol string) string {
	return filepath.Join(adi, "audit:"+strconv.Itoa(int(chainId))+":"+action+":"+symbol)
}

// Generate pending chain in format {account}#pending
func GeneratePendingChain(account string) string {
	return filepath.Join(account + "#pending")
}

// Generate data entry in format {entryhash}@{account}
func GenerateDataEntry(account string, entryhash string) string {
	return filepath.Join(entryhash + "@" + account)
}

func byte32(s []byte) (a *[32]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}
