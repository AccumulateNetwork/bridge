package fees

import (
	"fmt"
	"math"

	"github.com/AccumulateNetwork/bridge/schema"
	"github.com/go-playground/validator/v10"
)

// Operation is a helper to apply fees
type Operation struct {
	Token *schema.Token `json:"token"`
	In    int64         `json:"in" validate:"gt=0"`
	Out   int64         `json:"out" validate:"gt=0"`
}

// Mint applies minting fees to input and updates output amount
func (o *Operation) Mint(fees *schema.BridgeFees) (*Operation, error) {

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

	// deduct mint tx cost
	var out float64
	out = float64(o.In) - o.Token.EVMMintTxCost*math.Pow10(int(o.Token.Precision))

	if out <= 0 {
		return nil, fmt.Errorf("evm mint tx cost the same or higher than input amount")
	}

	// apply ratio and fees
	// fees are in bps (100 bps = 1%, 10000 = 100%)
	ratio := getRatio(o.Token.Precision, o.Token.EVMDecimals)
	out *= ratio * (10000 - float64(fees.MintFee)) / 10000

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
