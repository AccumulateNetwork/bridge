package evm

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestInfura(t *testing.T) {

	testPrivKey := "8a3aebfde003407a2a2d2fbb5db5627446dcd3b25f7e6b569540c928c4d45f78"

	c := &EVMClient{}

	c, err := c.ImportPrivateKey(testPrivKey)
	assert.NoError(t, err)

	conf := &config.Config{}
	conf.EVM.Node = "https://rinkeby.infura.io/v3/727aba752ecd48b79a6d508448cfa8aa"

	client, err := NewInfuraClient(conf)

	account := common.HexToAddress("0x221fB65CdB12Cc5eC0f9a2AfEe52D6c5CeF2B8bb")

	balance, err := client.Client.BalanceAt(context.Background(), account, nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(balance)

}
