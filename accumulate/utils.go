package accumulate

import (
	"strconv"
	"time"
	"unsafe"
)

// Generate bridge token account in format {chainId}-{symbol}
func GenerateTokenAccount(adi string, chainId int64, symbol string) string {

	return adi + "/" + strconv.Itoa(int(chainId)) + "-" + symbol

}

func nonceFromTimeNow() uint64 {
	t := time.Now()
	return uint64(t.Unix()*1e6) + uint64(t.Nanosecond())/1e3
}

func byte32(s []byte) (a *[32]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}
