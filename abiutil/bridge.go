package abiutil

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/gommon/log"
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

// UnpackBurnTxInputData unpacks bridge tx input data in hex format (without 0x)
func UnpackBurnTxInputData(data string) (*BurnData, error) {

	// load contract ABI
	abi, err := NewABI([]byte(BRIDGE_ABI))
	if err != nil {
		return nil, err
	}

	// decode txInput method signature
	decodedSig, err := hex.DecodeString(data[0:8])
	if err != nil {
		return nil, err
	}

	// recover Method from signature and ABI
	method, err := abi.MethodById(decodedSig)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// decode txInput Payload
	decodedData, err := hex.DecodeString(data[8:])
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})

	err = method.Inputs.UnpackIntoMap(m, decodedData)
	if err != nil {
		return nil, err
	}

	jsonInputs, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	burnData := &BurnData{}
	err = json.Unmarshal(jsonInputs, burnData)
	if err != nil {
		return nil, err
	}

	return burnData, nil

}
