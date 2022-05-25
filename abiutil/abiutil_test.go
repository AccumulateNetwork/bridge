package abiutil

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestNewABI(t *testing.T) {

	want := "0xa9059cbb000000000000000000000000c6386b0a95b60bcea480c876e3b1f9adb5b853140000000000000000000000000000000000000000000000000de0b6b3a7640000"

	abi, err := NewABI("abitest.abi")
	assert.NoError(t, err)

	method := "transfer"

	address := common.HexToAddress("0xC6386B0A95b60bCEa480C876e3b1F9AdB5B85314")

	amount := &big.Int{}
	amount.SetInt64(1e18)

	data, err := abi.Pack(method, address, amount)
	assert.NoError(t, err)

	got := hexutil.Encode(data)

	assert.Equal(t, got, want)

}
