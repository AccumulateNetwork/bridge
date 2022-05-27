package gnosis

import (
	"testing"

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
