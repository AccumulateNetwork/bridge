package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/global"
	"github.com/AccumulateNetwork/bridge/schema"
	"github.com/labstack/gommon/log"

	acmeurl "github.com/AccumulateNetwork/bridge/url"
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

func ValidateBurnEntry(entry *schema.BurnEvent, tx *abiutil.BurnData) error {

	log.Debug("Validating burn entry")

	log.Debug("entry amount=", entry.Amount, ", tx amount=", tx.Amount)
	if entry.Amount != tx.Amount.Int64() {
		return fmt.Errorf("entry amount=%d, tx amount=%d", entry.Amount, tx.Amount)
	}

	entryDestination, err := acmeurl.Parse(entry.Destination)
	if err != nil {
		return err
	}

	txDestination, err := acmeurl.Parse(tx.Destination)
	if err != nil {
		return err
	}

	log.Debug("entry destination=", entryDestination, ", tx destination=", txDestination)
	if entryDestination != txDestination {
		return fmt.Errorf("entry destination=%s, tx destination=%s", entry.Destination, tx.Destination)
	}

	log.Debug("entry token=", entry.TokenAddress, ", tx token=", tx.Token.Hex())
	// case insensitive comparison
	if !strings.EqualFold(entry.TokenAddress, tx.Token.Hex()) {
		return fmt.Errorf("entry destination=%s, tx destination=%s", entry.Destination, tx.Destination)
	}

	return nil

}

func ValidateReleaseTx(releaseTx *accumulate.TokenTx, tx *abiutil.BurnData) error {

	log.Debug("Validating release tx")

	if len(releaseTx.To) != 1 {
		return fmt.Errorf("expected 1 receiver (tx.Data.To), received=%d", len(releaseTx.To))
	}

	releaseTxAmount, err := strconv.ParseInt(releaseTx.To[0].Amount, 10, 64)
	if err != nil {
		return err
	}

	log.Debug("release tx amount=", releaseTxAmount, ", tx amount=", tx.Amount)
	if releaseTxAmount != tx.Amount.Int64() {
		return fmt.Errorf("release tx amount=%d, tx amount=%d", releaseTxAmount, tx.Amount)
	}

	releaseTxTo, err := acmeurl.Parse(releaseTx.To[0].URL)
	if err != nil {
		return err
	}

	txDestination, err := acmeurl.Parse(tx.Destination)
	if err != nil {
		return err
	}

	log.Debug("release tx destination=", releaseTx.To[0].URL, ", tx destination=", tx.Destination)
	if releaseTxTo == txDestination {
		return fmt.Errorf("entry destination=%s, tx destination=%s", releaseTx.To[0].URL, tx.Destination)
	}

	return nil

}
