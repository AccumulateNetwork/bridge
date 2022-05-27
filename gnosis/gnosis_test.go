package gnosis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportPrivateKey(t *testing.T) {

	testPrivKey := "19d2d4fee6210a01994379820053bcd09d0a242e030782049393ba1fb43f8d20"
	wantPublicKey := "0xBdBe86958C04183D63AfEaa9F362726E7eFB4A80"

	g := &Gnosis{}

	g, err := g.ImportPrivateKey(testPrivKey)
	assert.NoError(t, err)

	assert.Equal(t, wantPublicKey, g.PublicKey.String())

}
