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
	g.API = GNOSIS_API_RINKEBY
	g.SafeAddress = "0x5Ca3ad054405Cbe88b0907131cF021f8d24A6291"

	resp, err := g.GetSafe()
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Version)

}

func TestCreateSafeMultisigTx(t *testing.T) {

	// init gnosis safe
	g := &Gnosis{}
	g.API = GNOSIS_API_RINKEBY
	g.SafeAddress = "0x5Ca3ad054405Cbe88b0907131cF021f8d24A6291"
	g.BridgeAddress = "0xf2C0E57e40B85e6e016b3E04cC3268741740268e"
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
