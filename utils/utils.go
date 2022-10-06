package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/fees"
	"github.com/AccumulateNetwork/bridge/global"
	"github.com/AccumulateNetwork/bridge/schema"
	"github.com/go-playground/validator/v10"
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
	if entryDestination.Authority != txDestination.Authority || entryDestination.Path != txDestination.Path {
		return fmt.Errorf("entry destination=%s, tx destination=%s", entry.Destination, tx.Destination)
	}

	log.Debug("entry token=", entry.TokenAddress, ", tx token=", tx.Token.Hex())
	// case insensitive comparison
	if !strings.EqualFold(entry.TokenAddress, tx.Token.Hex()) {
		return fmt.Errorf("entry token=%s, tx token=%s", entry.TokenAddress, tx.Token.Hex())
	}

	return nil

}

func ValidateReleaseTx(releaseTx *accumulate.TokenTx, tx *abiutil.BurnData) error {

	// find token
	token := SearchEVMToken(tx.Token.String())

	if token == nil {
		return fmt.Errorf("token address %s is not supported by bridge", tx.Token.String())
	}

	operation := &fees.Operation{
		Token:  token,
		Amount: tx.Amount.Int64(),
	}

	outAmount, err := operation.ApplyFees(&global.BridgeFees, fees.OP_RELEASE)
	if err != nil {
		return err
	}

	if len(releaseTx.To) != 1 {
		return fmt.Errorf("expected 1 receiver (tx.Data.To), received=%d", len(releaseTx.To))
	}

	releaseTxAmount, err := strconv.ParseInt(releaseTx.To[0].Amount, 10, 64)
	if err != nil {
		return err
	}

	log.Debug("release tx amount=", releaseTxAmount, ", out tx amount=", outAmount)
	if releaseTxAmount != outAmount {
		return fmt.Errorf("release tx amount=%d, out tx amount=%d", releaseTxAmount, outAmount)
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
	if releaseTxTo.Authority != txDestination.Authority || releaseTxTo.Path != txDestination.Path {
		return fmt.Errorf("entry destination=%s, tx destination=%s", releaseTx.To[0].URL, tx.Destination)
	}

	return nil

}

func ValidateDepositTx(depositTx *accumulate.QueryTokenTxResponse) error {

	if depositTx.Type != accumulate.TX_TYPE_SYNTH_TOKEN_DEPOSIT {
		return fmt.Errorf("expected tx type %s, got %s", accumulate.TX_TYPE_SYNTH_TOKEN_DEPOSIT, depositTx.Type)
	}

	if depositTx.Data.Cause == "" {
		return fmt.Errorf("got empty cause")
	}

	if depositTx.Data.IsRefund {
		return fmt.Errorf("got refund tx")
	}

	return nil

}

func ValidateCauseTx(causeTx *accumulate.QueryTokenTxResponse) error {

	if causeTx.Type != accumulate.TX_TYPE_SEND_TOKENS {
		return fmt.Errorf("expected tx type %s, got %s", accumulate.TX_TYPE_SEND_TOKENS, causeTx.Type)
	}

	if causeTx.Transaction.Header.Memo == "" {
		return fmt.Errorf("no memo found")
	}

	return nil

}

func ValidateMintEntry(entry *schema.DepositEvent, tx *accumulate.QueryTokenTxResponse, cause *accumulate.QueryTokenTxResponse) error {

	log.Debug("Validating mint entry")

	amount, err := strconv.ParseInt(tx.Data.Amount, 10, 64)
	if err != nil {
		return err
	}

	log.Debug("entry amount=", entry.Amount, ", tx amount=", amount)
	if entry.Amount != amount {
		return fmt.Errorf("entry amount=%d, tx amount=%d", entry.Amount, amount)
	}

	log.Debug("entry token=", entry.TokenURL, ", tx token=", tx.Data.Token)
	// case insensitive comparison
	if !strings.EqualFold(entry.TokenURL, tx.Data.Token) {
		return fmt.Errorf("entry token=%s, tx token=%s", entry.TokenURL, tx.Data.Token)
	}

	log.Debug("entry destination=", entry.Destination, ", cause memo=", cause.Transaction.Header.Memo)
	// case insensitive comparison
	if !strings.EqualFold(entry.Destination, cause.Transaction.Header.Memo) {
		return fmt.Errorf("entry destination=%s, cause memo=%s", entry.Destination, cause.Transaction.Header.Memo)
	}

	// validate destination address
	validate := validator.New()
	err = validate.Var(entry.Destination, "required,eth_addr")
	if err != nil {
		return err
	}

	return nil

}
