package gnosis

import (
	"math/big"
	"testing"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestGetSafe(t *testing.T) {

	g := &Gnosis{}
	g.API = GNOSIS_API_GOERLI
	g.SafeAddress = "0x24BbA5D6fD7fC2Cbc293FDa6721c9BE6756D177a"

	resp, err := g.GetSafe()
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Version)

}

func TestCreateSafeMultisigTx(t *testing.T) {

	// init gnosis safe
	g := &Gnosis{}
	g.API = GNOSIS_API_GOERLI
	g.SafeAddress = "0x24BbA5D6fD7fC2Cbc293FDa6721c9BE6756D177a"
	g.BridgeAddress = "0x903f0dA0697FC1c81ecACc83b2A7445F392399e8"
	g.ImportPrivateKey("08108aadbbe82e9ffa1eba54158e7aacbec6115156242558ae7e594037220e4a")

	// get nonce
	safe, err := g.GetSafe()
	assert.NotZero(t, safe.Nonce)
	assert.NoError(t, err)

	// generate mint tx
	token := "0x4E780D102AADECF1BdC06d91542cf91960538a2D"
	recipient := "0xC6386B0A95b60bCEa480C876e3b1F9AdB5B85314"
	amount := big.NewInt(1e8)

	data, err := abiutil.GenerateMintTxData(token, recipient, amount)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// generate API request
	req := &NewMultisigTx{}
	req.To = g.BridgeAddress
	req.Data = hexutil.Encode(data)
	req.GasToken = abiutil.ZERO_ADDR
	req.RefundReceiver = abiutil.ZERO_ADDR
	req.Nonce = safe.Nonce

	resp, err := g.GetSafe()
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Version)

}
