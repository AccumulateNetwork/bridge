package abiutil

import (
	"bytes"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const ZERO_ADDR = "0x0000000000000000000000000000000000000000"

// NewABI construct ABI from []byte
func NewABI(abiBytes []byte) (*abi.ABI, error) {

	abiReader := bytes.NewReader(abiBytes)

	abi, err := abi.JSON(abiReader)
	if err != nil {
		return nil, err
	}

	return &abi, nil

}
