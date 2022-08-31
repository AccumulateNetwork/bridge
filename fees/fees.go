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
	Token  *schema.Token `json:"token"`
	Amount int64         `json:"amount" validate:"gt=0"`
}

// Mint applies minting fees to input and updates output amount
func (o *Operation) ApplyFees(fees *schema.BridgeFees, operation string) (int64, error) {

	var err error

	// init validator
	validate := validator.New()

	// validate fees
	err = validate.Struct(fees)
	if err != nil {
		return 0, err
	}

	// validate token
	err = validate.Struct(o.Token)
	if err != nil {
		return 0, err
	}

	var out float64
	var ratio float64
	var feeBps float64

	switch operation {
	case OP_MINT:
		out = float64(o.Amount) - o.Token.EVMMintTxCost*math.Pow10(int(o.Token.Precision))
		ratio = getRatio(o.Token.Precision, o.Token.EVMDecimals)
		feeBps = float64(fees.MintFee)
	case OP_RELEASE:
		out = float64(o.Amount)
		ratio = getRatio(o.Token.EVMDecimals, o.Token.Precision)
		feeBps = float64(fees.BurnFee)
	default:
		return 0, fmt.Errorf("invalid operation")
	}

	// apply ratio and fees
	// fees are in bps (100 bps = 1%, 10000 = 100%)
	out *= ratio * (10000 - feeBps) / 10000

	err = validate.Struct(o)
	if err != nil {
		return 0, err
	}

	res := int64(math.Floor(out))
	if res <= 0 {
		return 0, fmt.Errorf("output should be higher than 0")
	}

	return res, nil

}

// calculate in/out ratio
func getRatio(decimalsIn, decimalsOut int64) float64 {

	return math.Pow10(-1*int(decimalsIn)) * math.Pow10(int(decimalsOut))

}
