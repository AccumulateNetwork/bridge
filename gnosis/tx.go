package gnosis

import (
	"fmt"
	"math/big"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core"
)

func (g *Gnosis) SignMintTx(tokenAddress string, recipientAddress string, amount *big.Int) ([]byte, []byte, error) {

	safe, err := g.GetSafe()
	if err != nil {
		return nil, nil, err
	}

	data, err := abiutil.GenerateMintTx(tokenAddress, recipientAddress, amount)
	if err != nil {
		return nil, nil, err
	}

	// get contract transaction hash
	gnosisSafeTx := core.GnosisSafeTx{
		Safe:           common.NewMixedcaseAddress(common.HexToAddress(g.SafeAddress)),
		To:             common.NewMixedcaseAddress(common.HexToAddress(g.BridgeAddress)),
		Value:          *math.NewDecimal256(0),
		GasPrice:       *math.NewDecimal256(0),
		Data:           (*hexutil.Bytes)(&data),
		Operation:      0,
		GasToken:       common.HexToAddress(abiutil.ZERO_ADDR),
		RefundReceiver: common.HexToAddress(abiutil.ZERO_ADDR),
		BaseGas:        *big.NewInt(0),
		SafeTxGas:      *big.NewInt(0),
		Nonce:          *big.NewInt(safe.Nonce),
		ChainId:        math.NewHexOrDecimal256(int64(g.ChainId)),
	}

	typedData := gnosisSafeTx.ToTypedData()

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, nil, err
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, nil, err
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	sighash := crypto.Keccak256Hash(rawData)

	contractTxHash, err := hexutil.Decode(sighash.Hex())
	if err != nil {
		return nil, nil, err
	}

	signature, err := crypto.Sign(contractTxHash, g.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	if signature[64] == 0 || signature[64] == 1 {
		signature[64] += 27
	}

	return contractTxHash, signature, nil

}
