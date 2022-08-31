package fees

import (
	"testing"

	"github.com/AccumulateNetwork/bridge/schema"
	"github.com/stretchr/testify/assert"
)

func TestMint(t *testing.T) {

	var err error

	fees := &schema.BridgeFees{}
	token := &schema.Token{}
	op := &Operation{Token: token}

	// TEST 1: negative mint fees
	fees.BurnFee = 0
	fees.MintFee = -1
	_, err = op.Mint(fees)
	assert.Error(t, err)

	// TEST 2: negative mint tx cost
	fees.MintFee = 0
	op.Token.EVMMintTxCost = -1
	_, err = op.Mint(fees)
	assert.Error(t, err)

	// TEST 3: amount <= mint tx cost
	fees.MintFee = 10 // 10 bps = 0.1%
	op.Token.EVMMintTxCost = 50
	op.Token.Precision = 8
	op.Token.EVMDecimals = 8
	op.In = 50 * 1e8
	_, err = op.Mint(fees)
	assert.Error(t, err)

	// TEST 4: calculate fees
	op.In = 1000 * 1e8
	op, err = op.Mint(fees)
	assert.NoError(t, err)
	// (1000 [in] - 50 [mint cost]) - 0.1% = 949.05
	assert.Equal(t, op.Out, int64(949.05*1e8))

	// TEST 5: float evm mint cost
	op.Token.EVMMintTxCost = 0.5
	op, err = op.Mint(fees)
	assert.NoError(t, err)
	// (1000 [in] - 0.5 [mint cost]) - 0.1% = 998.5005
	assert.Equal(t, op.Out, int64(998.5005*1e8))

	// TEST 6: rounding down
	op.Token.EVMDecimals = 0
	op, err = op.Mint(fees)
	assert.NoError(t, err)
	// (1000 [in] - 0.5 [mint cost]) - 0.1% = 998.5005 = 998 (rounding down)
	assert.Equal(t, op.Out, int64(998))

	// TEST 7: zero out
	op.In = 51 * 1e8
	op.Token.EVMMintTxCost = 50
	_, err = op.Mint(fees)
	// (51 [in] - 50 [mint cost]) - 0.1% = 0.999 = 0 (rounding down)
	assert.Error(t, err)

}

func TestGetRatio(t *testing.T) {

	var ratio float64
	var expected float64

	ratio = getRatio(8, 8)
	expected = 1
	assert.Equal(t, ratio, expected)

	ratio = getRatio(3, 0)
	expected = 1e-3
	assert.Equal(t, ratio, expected)

	ratio = getRatio(0, 8)
	expected = 1e8
	assert.Equal(t, ratio, expected)

}
