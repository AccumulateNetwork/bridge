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
		Sender:         common.NewMixedcaseAddress(common.HexToAddress(g.PublicKey.Hex())),
		Safe:           common.NewMixedcaseAddress(common.HexToAddress(g.SafeAddress)),
		To:             common.NewMixedcaseAddress(common.HexToAddress(g.BridgeAddress)),
		Value:          *math.NewDecimal256(0),
		GasPrice:       *math.NewDecimal256(0),
		Data:           (*hexutil.Bytes)(&data),
		Operation:      0,
		GasToken:       common.HexToAddress(ZERO_ADDR),
		RefundReceiver: common.HexToAddress(ZERO_ADDR),
		BaseGas:        *common.Big0,
		SafeTxGas:      *big.NewInt(0),
		Nonce:          *big.NewInt(safe.Nonce),
	}

	typedData := gnosisSafeTx.ToTypedData()

	domainHash, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, nil, err
	}
	primaryTypeHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, nil, err
	}

	encodedTx := []byte{1, 19}
	encodedTx = append(encodedTx, domainHash...)
	encodedTx = append(encodedTx, primaryTypeHash...)

	encodedTxHash := crypto.Keccak256Hash(encodedTx)

	fmt.Println("encodedTxHash:", encodedTxHash.Hex())

	contractTxHash, err := hexutil.Decode(encodedTxHash.Hex())
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
