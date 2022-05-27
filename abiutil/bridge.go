package abiutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// GenerateMintTx generates hex-encoded string, containing the instruction for gnosis safe to mint token via bridge
func GenerateMintTx(tokenAddress string, recipientAddress string, amount *big.Int) (string, error) {

	abi, err := NewABI("bridge.abi")
	if err != nil {
		return "", err
	}

	method := "mint"
	token := common.HexToAddress(tokenAddress)
	recipient := common.HexToAddress(recipientAddress)

	data, err := abi.Pack(method, token, recipient, amount)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(data), nil

}
