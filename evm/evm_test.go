package evm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportPrivateKey(t *testing.T) {

	testPrivKey := "19d2d4fee6210a01994379820053bcd09d0a242e030782049393ba1fb43f8d20"
	wantPublicKey := "0xBdBe86958C04183D63AfEaa9F362726E7eFB4A80"

	c := &EVMClient{}

	c, err := c.ImportPrivateKey(testPrivKey)
	assert.NoError(t, err)

	assert.Equal(t, wantPublicKey, c.PublicKey.String())

}

/*
func TestNewEVMClient(t *testing.T) {

	testPrivKey := "8a3aebfde003407a2a2d2fbb5db5627446dcd3b25f7e6b569540c928c4d45f78"

	c := &EVMClient{}

	c, err := c.ImportPrivateKey(testPrivKey)
	assert.NoError(t, err)

	conf := &config.Config{}
	conf.EVM.InfuraProjectID = "727aba752ecd48b79a6d508448cfa8aa"
	conf.EVM.ChainId = 4

	client, err := NewEVMClient(conf)
	assert.NoError(t, err)

		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID: conf.EVM.ChainId,
		})

		// account := common.HexToAddress("0x221fB65CdB12Cc5eC0f9a2AfEe52D6c5CeF2B8bb")

		signedTx, err := types.SignTx(tx, types.NewLondonSigner(big.NewInt(4)), c.PrivateKey)
		ts := types.Transactions{signedTx}
		rawTxBytes, _ := rlp.EncodeToBytes(ts[0])
		rawTxHex := hex.EncodeToString(rawTxBytes)

		tx1 := new(types.Transaction)
		rlp.DecodeBytes(rawTxBytes, &tx)

		err = client.Client.SendTransaction(context.Background(), tx1)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("tx sent: %s", tx.Hash().Hex()) // tx sent: 0xc429e5f128387d224ba8bed6885e86525e14bfdc2eb24b5e9c3351a1176fd81f

		fmt.Printf(rawTxHex)

}
*/
