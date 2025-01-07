package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (e *EVMClient) Submit(gasPrice float64, priorityFee float64, to *common.Address, value int64, data []byte) (*types.Transaction, error) {
	// Determine which transaction type to create based on chain ID
	if e.ChainId == 56 {
		return e.submitLegacyTx(gasPrice, to, value, data)
	}
	return e.submitEIP1559Tx(gasPrice, priorityFee, to, value, data)
}

func (e *EVMClient) submitLegacyTx(gasPrice float64, to *common.Address, value int64, data []byte) (*types.Transaction, error) {
	chainId := big.NewInt(int64(e.ChainId))

	// Calculate gas price
	gasFeeCap := new(big.Int)
	gasFeeCap.SetString(fmt.Sprintf("%.0f", gasPrice*1e9), 10)

	txValue := big.NewInt(value)

	// Get nonce
	fromNonce, err := e.Client.PendingNonceAt(context.Background(), e.PublicKey)
	if err != nil {
		fmt.Print("can not get nonce: ")
		return nil, err
	}

	// Create legacy transaction
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    fromNonce,
		GasPrice: gasFeeCap,
		Gas:      uint64(e.GasLimit),
		To:       to,
		Value:    txValue,
		Data:     data,
	})

	// Sign and send transaction
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainId), e.PrivateKey)
	if err != nil {
		fmt.Print("can not sign tx: ")
		return nil, err
	}

	err = e.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Print("can not send tx: ")
		return nil, err
	}

	return signedTx, nil
}

func (e *EVMClient) submitEIP1559Tx(gasPrice float64, priorityFee float64, to *common.Address, value int64, data []byte) (*types.Transaction, error) {
	chainId := big.NewInt(int64(e.ChainId))

	// Calculate gas fees
	gasFeeCap := new(big.Int)
	gasFeeCap.SetString(fmt.Sprintf("%.0f", gasPrice*1e9), 10)

	gasTipCap := new(big.Int)
	gasTipCap.SetString(fmt.Sprintf("%.0f", priorityFee*1e9), 10)

	txValue := big.NewInt(value)

	// Get nonce
	fromNonce, err := e.Client.PendingNonceAt(context.Background(), e.PublicKey)
	if err != nil {
		fmt.Print("can not get nonce: ")
		return nil, err
	}

	// Create EIP-1559 transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     fromNonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       uint64(e.GasLimit),
		To:        to,
		Value:     txValue,
		Data:      data,
	})

	// Sign and send transaction
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainId), e.PrivateKey)
	if err != nil {
		fmt.Print("can not sign tx: ")
		return nil, err
	}

	err = e.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Print("can not send tx: ")
		return nil, err
	}

	return signedTx, nil
}
