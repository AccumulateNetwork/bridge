package evm

import (
	"context"
	"fmt"

	"github.com/AccumulateNetwork/bridge/binding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ERC20 struct {
	Address  string
	Symbol   string
	Decimals int64
	Owner    string
}

type Tx struct {
	TxHash  string
	ChainId int64
	To      string
	Data    []byte
}

// GetERC20 gets ERC20 Token info
func (e *EVMClient) GetERC20(tokenAddress string) (*ERC20, error) {

	token := &ERC20{}

	address := common.HexToAddress(tokenAddress)

	instance, err := binding.NewWrappedToken(address, e.Client)
	if err != nil {
		return nil, err
	}

	owner, err := instance.Owner(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	token.Owner = owner.String()
	token.Symbol = symbol
	token.Decimals = int64(decimals)

	return token, nil

}

// GetTx gets tx by hash
func (e *EVMClient) GetTx(hash string) (*Tx, error) {

	tx := &Tx{}
	txid := common.HexToHash(hash)

	evmTx, isPending, err := e.Client.TransactionByHash(context.Background(), txid)
	if err != nil {
		return nil, err
	}

	// check if tx is pending
	if isPending {
		return nil, fmt.Errorf("tx %s is pending, skipping", evmTx.Hash())
	}

	tx.TxHash = evmTx.Hash().Hex()

	// check chain id
	if int64(e.ChainId) != evmTx.ChainId().Int64() {
		return nil, fmt.Errorf("received chainId %d, expected %d", evmTx.ChainId(), e.ChainId)
	}

	tx.ChainId = evmTx.ChainId().Int64()

	// fill additional data to check it in main module
	tx.To = evmTx.To().Hex()
	tx.Data = evmTx.Data()

	return tx, nil

}
