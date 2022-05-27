package gnosis

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestPrivatePublicKey(t *testing.T) {

	testPrivKey := "19d2d4fee6210a01994379820053bcd09d0a242e030782049393ba1fb43f8d20"
	wantPublicKey := "0xBdBe86958C04183D63AfEaa9F362726E7eFB4A80"

	g := &Gnosis{}

	privateKey, err := crypto.HexToECDSA(testPrivKey)
	assert.NoError(t, err)

	g.PrivateKey = privateKey

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	assert.True(t, ok)

	g.PublicKey = crypto.PubkeyToAddress(*publicKeyECDSA)

	assert.Equal(t, wantPublicKey, g.PublicKey.String())

}
