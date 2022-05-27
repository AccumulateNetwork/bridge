package abiutil

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMintTx(t *testing.T) {

	want := "0xc6c3bbe60000000000000000000000004e780d102aadecf1bdc06d91542cf91960538a2d000000000000000000000000c6386b0a95b60bcea480c876e3b1f9adb5b853140000000000000000000000000000000000000000000000000000000005f5e100"

	token := "0x4E780D102AADECF1BdC06d91542cf91960538a2D"
	recipient := "0xC6386B0A95b60bCEa480C876e3b1F9AdB5B85314"
	amount := &big.Int{}
	amount.SetInt64(1e8)

	got, err := GenerateMintTx(token, recipient, amount)
	assert.NoError(t, err)

	assert.Equal(t, want, got)

}
