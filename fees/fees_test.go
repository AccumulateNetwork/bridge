package fees

import (
	"testing"

	"github.com/AccumulateNetwork/bridge/schema"
	"github.com/stretchr/testify/assert"
)

func TestApplyFees(t *testing.T) {

	var err error
	var out int64

	fees := &schema.BridgeFees{}
	token := &schema.Token{}
	op := &Operation{Token: token}

	// TEST 1: negative mint fees
	fees.BurnFee = 0
	fees.MintFee = -1
	_, err = op.ApplyFees(fees, OP_MINT)
	assert.Error(t, err)

	// TEST 2: negative mint tx cost
	fees.MintFee = 0
	op.Token.EVMMintTxCost = -1
	_, err = op.ApplyFees(fees, OP_MINT)
	assert.Error(t, err)

	// TEST 3: amount <= mint tx cost
	fees.MintFee = 10 // 10 bps = 0.1%
	op.Token.EVMMintTxCost = 50
	op.Token.Precision = 8
	op.Token.EVMDecimals = 8
	op.Amount = 50 * 1e8
	_, err = op.ApplyFees(fees, OP_MINT)
	assert.Error(t, err)

	// TEST 4: calculate fees
	op.Amount = 1000 * 1e8
	out, err = op.ApplyFees(fees, OP_MINT)
	assert.NoError(t, err)
	// 1000 [in] - 0.1% - 50 [mint cost] = 949
	assert.Equal(t, out, int64(949*1e8))

	// TEST 5: float evm mint cost
	op.Token.EVMMintTxCost = 0.5
	out, err = op.ApplyFees(fees, OP_MINT)
	assert.NoError(t, err)
	// 1000 [in] - 0.1% - 0.5 [mint cost] = 998.5
	assert.Equal(t, out, int64(998.5*1e8))

	// TEST 6: rounding down
	token2 := &schema.Token{Precision: 1, EVMDecimals: 1, EVMMintTxCost: 0.5}
	op.Token = token2
	op.Amount = 5 * 10
	fees.MintFee = 1
	out, err = op.ApplyFees(fees, OP_MINT)
	assert.NoError(t, err)
	// 5 [in] - 0.01% - 0.5 [mint cost] = 4.4995 = 4.4 (rounding down)
	assert.Equal(t, out, int64(4.4*10))

	// TEST 7: zero out
	op.Amount = 51 * 10
	fees.MintFee = 10
	op.Token.EVMMintTxCost = 50.9
	_, err = op.ApplyFees(fees, OP_MINT)
	// 51 [in] - 0.1% - 50.9 [mint cost] = 0.0949 = 0 (rounding down)
	assert.Error(t, err)

	// TEST 8: burn-release (zero fee)
	token3 := &schema.Token{Precision: 8, EVMDecimals: 8, EVMMintTxCost: 1}
	op.Token = token3
	op.Amount = 51 * 1e8
	out, err = op.ApplyFees(fees, OP_RELEASE)
	assert.NoError(t, err)
	// 51 [in] = 51
	assert.Equal(t, out, int64(51*1e8))

	// TEST 9: burn-release (non-zero fee)
	fees.BurnFee = 1000 // 1000 bps = 10%
	out, err = op.ApplyFees(fees, OP_RELEASE)
	assert.NoError(t, err)
	// 51 [in] - 10% = 45.9
	assert.Equal(t, out, int64(45.9*1e8))

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
