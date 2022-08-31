package fees

import (
	"fmt"
	"math"

	"github.com/AccumulateNetwork/bridge/schema"
	"github.com/go-playground/validator/v10"
)

const OP_MINT = "mint"
const OP_RELEASE = "release"

// Operation is a helper to apply fees
type Operation struct {
	Token *schema.Token `json:"token"`
	In    int64         `json:"in" validate:"gt=0"`
	Out   int64         `json:"out" validate:"gt=0"`
}

// Mint applies minting fees to input and updates output amount
func (o *Operation) ApplyFees(fees *schema.BridgeFees, operation string) (*Operation, error) {

	var err error

	// init validator
	validate := validator.New()

	// validate fees
	err = validate.Struct(fees)
	if err != nil {
		return nil, err
	}

	// validate token
	err = validate.Struct(o.Token)
	if err != nil {
		return nil, err
	}

	var out float64
	var ratio float64
	var feeBps float64

	switch operation {
	case OP_MINT:
		out = float64(o.In) - o.Token.EVMMintTxCost*math.Pow10(int(o.Token.Precision))
		ratio = getRatio(o.Token.Precision, o.Token.EVMDecimals)
		feeBps = float64(fees.MintFee)
	case OP_RELEASE:
		out = float64(o.In)
		ratio = getRatio(o.Token.EVMDecimals, o.Token.Precision)
		feeBps = float64(fees.BurnFee)
	default:
		return nil, fmt.Errorf("invalid operation")
	}

	if out <= 0 {
		return nil, fmt.Errorf("evm mint tx cost the same or higher than input amount")
	}

	// apply ratio and fees
	// fees are in bps (100 bps = 1%, 10000 = 100%)
	out *= ratio * (10000 - feeBps) / 10000
	o.Out = int64(math.Floor(out))

	err = validate.Struct(o)
	if err != nil {
		return nil, err
	}

	return o, nil

}

// calculate in/out ratio
func getRatio(decimalsIn, decimalsOut int64) float64 {

	return math.Pow10(-1*int(decimalsIn)) * math.Pow10(int(decimalsOut))

}
