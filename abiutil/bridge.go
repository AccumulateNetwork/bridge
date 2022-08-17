package abiutil

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BurnData struct {
	Token       common.Address
	Destination string
	Amount      *big.Int
}

const BRIDGE_ABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contract WrappedToken\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"destination\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contract WrappedToken\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contract WrappedToken\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"RenounceTokenOwnership\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contract WrappedToken\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"TransferTokenOwnership\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"contract WrappedToken\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"destination\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contract WrappedToken\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contract WrappedToken\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"renounceTokenOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contract WrappedToken\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferTokenOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// GenerateMintTx generates hex-encoded string, containing the instruction for gnosis safe to mint token via bridge
func GenerateMintTxData(tokenAddress string, recipientAddress string, amount *big.Int) ([]byte, error) {

	abi, err := NewABI([]byte(BRIDGE_ABI))
	if err != nil {
		return nil, err
	}

	method := "mint"
	token := common.HexToAddress(tokenAddress)
	recipient := common.HexToAddress(recipientAddress)

	data, err := abi.Pack(method, token, recipient, amount)
	if err != nil {
		return nil, err
	}

	return data, nil

}

// UnpackBurnTxInputData unpacks bridge tx input data
func UnpackBurnTxInputData(data []byte) (*BurnData, error) {

	abi, err := NewABI([]byte(BRIDGE_ABI))
	if err != nil {
		return nil, err
	}

	method, ok := abi.Methods["burn"]
	if !ok {
		return nil, fmt.Errorf("error finding method burn")
	}

	var v map[string]interface{}

	// unpack method inputs
	err = method.Inputs.UnpackIntoMap(v, data[4:])
	if err != nil {
		return nil, err
	}

	//fmt.Println(v["amount"].(*big.Int))

	burnData := &BurnData{}

	return burnData, nil

}
