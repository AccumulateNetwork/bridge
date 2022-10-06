package abiutil

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestGenerateMintTx(t *testing.T) {

	want := "0xc6c3bbe60000000000000000000000004e780d102aadecf1bdc06d91542cf91960538a2d000000000000000000000000c6386b0a95b60bcea480c876e3b1f9adb5b853140000000000000000000000000000000000000000000000000000000005f5e100"

	token := "0x4E780D102AADECF1BdC06d91542cf91960538a2D"
	recipient := "0xC6386B0A95b60bCEa480C876e3b1F9AdB5B85314"
	amount := &big.Int{}
	amount.SetInt64(1e8)

	got, err := GenerateMintTxData(token, recipient, amount)
	assert.NoError(t, err)

	assert.Equal(t, want, hexutil.Encode(got))

}

func TestUnpackTxInputData(t *testing.T) {

	token := "0xe3fA338f248d640bF759E2a89283503a2281612a"
	dest := "acc://abdafe3eb60d205905e10e5a2129e9567292646b968ecb7b/ACME"
	amount := &big.Int{}
	amount.SetInt64(10e8)

	inputData := "c45b71de000000000000000000000000e3fa338f248d640bf759e2a89283503a2281612a0000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000003b9aca00000000000000000000000000000000000000000000000000000000000000003b6163633a2f2f6162646166653365623630643230353930356531306535613231323965393536373239323634366239363865636237622f41434d450000000000"

	unpacked, err := UnpackBurnTxInputData(inputData)
	assert.NoError(t, err)
	assert.Equal(t, unpacked.Token.Hex(), token)
	assert.Equal(t, unpacked.Destination, dest)
	assert.Equal(t, unpacked.Amount, amount)

}
