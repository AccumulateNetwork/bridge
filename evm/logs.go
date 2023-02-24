package evm

import (
	"context"
	"math/big"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/gommon/log"
)

type BlockRange struct {
	From int64
	To   int64
}

type EventLog struct {
	TxID        common.Hash
	BlockHeight uint64
	Token       common.Address
	Amount      *big.Int
	To          common.Address
	Destination string
}

// ParseEventLog parses event logs from Ethereum
func (e *EVMClient) ParseBridgeLogs(eventName string, bridgeAddress string, blocks *BlockRange) ([]*EventLog, error) {

	events := []*EventLog{}

	contractAbi, err := abiutil.NewABI([]byte(abiutil.BRIDGE_ABI))
	if err != nil {
		return nil, err
	}

	// calculate event hash from event name
	eventSig := []byte(contractAbi.Events[eventName].Sig)
	eventHash := crypto.Keccak256Hash(eventSig)

	// prepate filters for query
	commonHash := []common.Hash{}
	commonHash = append(commonHash, eventHash)
	address := common.HexToAddress(bridgeAddress)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			address,
		},
		Topics: [][]common.Hash{
			commonHash,
		},
	}

	if blocks.From > 0 {
		query.FromBlock = big.NewInt(blocks.From)
	}

	if blocks.To > 0 {
		query.ToBlock = big.NewInt(blocks.To)
	}

	logs, err := e.Client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	for _, vLog := range logs {
		event := &EventLog{}

		err := contractAbi.UnpackIntoInterface(event, eventName, vLog.Data)
		if err != nil {
			log.Error(err)
			continue
		}

		event.TxID = vLog.TxHash
		event.BlockHeight = vLog.BlockNumber

		events = append(events, event)
	}

	return events, nil

}
