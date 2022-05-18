package gnosis

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/AccumulateNetwork/bridge/config"

	"github.com/labstack/gommon/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core"
)

const (
	ZERO_ADDR          = "0x0000000000000000000000000000000000000000"
	GNOSIS_API_MAINNET = "https://safe-transaction.gnosis.io/api/v1/"
	GNOSIS_API_RINKEBY = "https://safe-transaction.rinkeby.gnosis.io/api/v1/"
)

type Gnosis struct {
	API           string
	SafeAddress   string
	BridgeAddress string
	PrivateKey    *ecdsa.PrivateKey
	PublicKey     common.Address
}

type ResponseSafe struct {
	Address         string   `json:"address"`
	Nonce           int64    `json:"nonce"`
	Threshold       int64    `json:"threshold"`
	Owners          []string `json:"owners"`
	MasterCopy      string   `json:"masterCopy"`
	Modules         []string `json:"modules"`
	FallbackHandler string   `json:"fallbackHandler"`
	Guard           string   `json:"guard"`
	Version         string   `json:"version"`
}

type RequestEstSafeTxGas struct {
	To        string `json:"to"`
	Value     int64  `json:"value"`
	Data      string `json:"data"`
	Operation int64  `json:"operation"`
}

type ResponseEstSafeTxGas struct {
	SafeTxGas int64 `json:"safeTxGas,string,omitempty"`
}

type RequestGnosisTx struct {
	To                      string  `json:"to"`
	Value                   int64   `json:"value"`
	Data                    *string `json:"data"`
	Operation               int64   `json:"operation"`
	GasToken                string  `json:"gasToken"`
	SafeTxGas               int64   `json:"safeTxGas"`
	BaseGas                 int64   `json:"baseGas"`
	GasPrice                int64   `json:"gasPrice"`
	RefundReceiver          string  `json:"refundReceiver"`
	Nonce                   int64   `json:"nonce"`
	ContractTransactionHash string  `json:"contractTransactionHash"`
	Sender                  string  `json:"sender"`
	Signature               string  `json:"signature"`
	Origin                  *string `json:"origin"`
}

type ResponseErrorGnosisTx struct {
	NonFieldErrors []string `json:"nonFieldErrors"`
}

// NewGnosis constructs the gnosis safe
func NewGnosis(conf *config.Config) (*Gnosis, error) {

	g := &Gnosis{}

	switch conf.EVM.ChainId {

	case 1:
		g.API = GNOSIS_API_MAINNET
	case 4:
		g.API = GNOSIS_API_RINKEBY
	default:
		return nil, fmt.Errorf("received unknown chainId from config: %s", strconv.Itoa(conf.EVM.ChainId))

	}

	if conf.EVM.SafeAddress == "" {
		return nil, fmt.Errorf("received empty safeAddress from config: %s", conf.EVM.SafeAddress)
	}
	g.SafeAddress = conf.EVM.SafeAddress

	if conf.EVM.BridgeAddress == "" {
		return nil, fmt.Errorf("received empty bridgeAddress from config: %s", conf.EVM.BridgeAddress)
	}
	g.BridgeAddress = conf.EVM.BridgeAddress

	if conf.EVM.PrivateKey == "" {
		return nil, fmt.Errorf("received empty privateKey from config: %s", conf.EVM.PrivateKey)
	}

	privateKey, err := crypto.HexToECDSA(conf.EVM.PrivateKey)
	if err != nil {
		return nil, err
	}

	g.PrivateKey = privateKey

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	g.PublicKey = crypto.PubkeyToAddress(*publicKeyECDSA)

	return g, nil

}

// GetSafe gets safe info and current nonce
func (g *Gnosis) GetSafe() (*ResponseSafe, error) {

	body, err := g.makeRequest("safes/"+g.SafeAddress, nil)
	if err != nil {
		return nil, err
	}

	var safe ResponseSafe

	if err = json.Unmarshal(body, &safe); err != nil {
		return nil, err
	}

	return &safe, nil

}

// EstimateSafeTxGas estimates safe tx costs
func (g *Gnosis) EstimateSafeTxGas(req *RequestEstSafeTxGas) (*ResponseEstSafeTxGas, error) {

	req.To = g.SafeAddress

	reqBytes, _ := json.Marshal(req)

	body, err := g.makeRequest("safes/"+g.SafeAddress+"/multisig-transactions/estimations/", reqBytes)
	if err != nil {
		return nil, err
	}

	var safe ResponseEstSafeTxGas

	if err = json.Unmarshal(body, &safe); err != nil {
		return nil, err
	}

	return &safe, nil

}

// MintTx calls mint function in Accumulate Bridge smart contract
func (g *Gnosis) MintTx(safeTxGas *int64, nonce *int64) (string, error) {

	// get contract transaction hash
	gnosisSafeTx := core.GnosisSafeTx{
		Sender:         common.NewMixedcaseAddress(g.PublicKey),
		Safe:           common.NewMixedcaseAddress(common.HexToAddress(g.SafeAddress)),
		To:             common.NewMixedcaseAddress(common.HexToAddress(g.BridgeAddress)),
		Value:          *math.NewDecimal256(0),
		GasPrice:       *math.NewDecimal256(0),
		Data:           &hexutil.Bytes{},
		Operation:      0,
		GasToken:       common.HexToAddress(ZERO_ADDR),
		RefundReceiver: common.HexToAddress(ZERO_ADDR),
		BaseGas:        *common.Big0,
		SafeTxGas:      *big.NewInt(*safeTxGas),
		Nonce:          *big.NewInt(*nonce),
	}

	typedData := gnosisSafeTx.ToTypedData()

	domainHash, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", err
	}
	primaryTypeHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", err
	}

	encodedTx := []byte{1, 19}
	encodedTx = append(encodedTx, domainHash...)
	encodedTx = append(encodedTx, primaryTypeHash...)

	encodedTxHash := crypto.Keccak256Hash(encodedTx)

	log.Info("EncodedTxHash: %s", encodedTxHash.Hex())

	signature, err := crypto.Sign(encodedTxHash.Bytes(), g.PrivateKey)
	if err != nil {
		return "", err
	}

	if signature[64] == 0 || signature[64] == 1 {
		signature[64] += 27
	}

	gnosisReq := RequestGnosisTx{
		To:                      g.BridgeAddress,
		Value:                   0,
		Data:                    nil,
		Operation:               0,
		GasToken:                ZERO_ADDR,
		SafeTxGas:               *safeTxGas,
		BaseGas:                 0,
		GasPrice:                0,
		RefundReceiver:          ZERO_ADDR,
		Nonce:                   *nonce,
		ContractTransactionHash: encodedTxHash.Hex(),
		Sender:                  g.PublicKey.String(),
		Signature:               hexutil.Encode(signature),
		Origin:                  nil,
	}

	log.Info(gnosisReq)

	req, err := json.Marshal(gnosisReq)
	if err != nil {
		return "", err
	}

	body, err := g.makeRequest("safes/"+g.SafeAddress+"/multisig-transactions/", req)
	if err != nil {
		return "", err
	}

	var respErr ResponseErrorGnosisTx

	if err = json.Unmarshal(body, &respErr); err != nil {
		return "", err
	}

	return strings.Join(respErr.NonFieldErrors, "\n"), nil

}

// internal
func (g *Gnosis) makeRequest(path string, req []byte) ([]byte, error) {

	var resp *http.Response
	var err error

	if req != nil {
		resp, err = http.Post(g.API+path, "application/json", bytes.NewBuffer(req))
	} else {
		resp, err = http.Get(g.API + path)
	}
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
