package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://rinkeby.infura.io/v3/727aba752ecd48b79a6d508448cfa8aa")

	if err != nil {
		log.Fatalln("Oops! There was a problem", err)
	} else {
		account := common.HexToAddress("0x221fB65CdB12Cc5eC0f9a2AfEe52D6c5CeF2B8bb")

		balance, err := client.BalanceAt(context.Background(), account, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(balance)
	}
}
