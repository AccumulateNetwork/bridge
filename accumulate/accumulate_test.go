package accumulate

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportPrivateKey(t *testing.T) {

	testPrivKey := "7d36bbb9f6c36bd4883095ae12795a85def0f3027332e1930fbd4626c8f8ac921fece78f587776b6f178cbb1437ff0102f039a0872ec89766da084be84221cc8"
	wantPublicKey := "1fece78f587776b6f178cbb1437ff0102f039a0872ec89766da084be84221cc8"

	wantPublicKeyBytes, err := hex.DecodeString(wantPublicKey)
	assert.NoError(t, err)

	c := &AccumulateClient{}

	c, err = c.ImportPrivateKey(testPrivKey)
	assert.NoError(t, err)

	assert.Equal(t, sha256.Sum256(wantPublicKeyBytes), sha256.Sum256(c.PublicKey))

}
