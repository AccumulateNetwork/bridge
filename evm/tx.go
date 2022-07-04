package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (e *EVMClient) SubmitEIP1559Tx(gasLimit int64, gasPrice int64, priorityFee int64, to *common.Address, value int64, data []byte) (*types.Transaction, error) {

	// convert to big.Int
	chainId := &big.Int{}
	chainId.SetInt64(int64(e.ChainId))

	gasFeeCap := &big.Int{}
	gasFeeCap.SetInt64(gasPrice * 1e9)

	gasTipCap := &big.Int{}
	gasTipCap.SetInt64(priorityFee * 1e9)

	txValue := &big.Int{}
	txValue.SetInt64(value)

	// nonce of tx sender
	fromNonce, err := e.Client.PendingNonceAt(context.Background(), e.PublicKey)
	if err != nil {
		fmt.Print("can not get nonce: ")
		return nil, err
	}

	// generate new tx EIP-1559
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     fromNonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       uint64(200000),
		To:        to,
		Value:     txValue,
		Data:      data,
	})

	// sign tx
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
