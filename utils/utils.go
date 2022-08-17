package utils

import (
	"strings"

	"github.com/AccumulateNetwork/bridge/global"
	"github.com/AccumulateNetwork/bridge/schema"
)

func SearchEVMToken(address string) *schema.Token {

	for _, t := range global.Tokens.Items {
		if strings.EqualFold(t.EVMAddress, address) {
			return t
		}
	}

	return nil

}

func SearchAccumulateToken(url string) *schema.Token {

	for _, t := range global.Tokens.Items {
		if strings.EqualFold(t.URL, url) {
			return t
		}
	}

	return nil

}
