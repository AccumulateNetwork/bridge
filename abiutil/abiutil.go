package abiutil

import (
	"bytes"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// NewABI construct ABI from JSON string
func NewABI(abiJSON string) (*abi.ABI, error) {

	abiBytes, err := ioutil.ReadFile(abiJSON)
	if err != nil {
		return nil, err
	}

	abiReader := bytes.NewReader(abiBytes)

	abi, err := abi.JSON(abiReader)
	if err != nil {
		return nil, err
	}

	return &abi, nil

}
