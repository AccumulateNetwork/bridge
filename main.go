package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AccumulateNetwork/bridge/abiutil"
	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/api"
	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/evm"
	"github.com/AccumulateNetwork/bridge/global"
	"github.com/AccumulateNetwork/bridge/gnosis"
	"github.com/AccumulateNetwork/bridge/schema"
	acmeurl "github.com/AccumulateNetwork/bridge/url"
	"github.com/AccumulateNetwork/bridge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
)

const LEADER_MIN_DURATION = 1
const NUMBER_OF_ACCUMULATE_TOKEN_TXS = 100
const NUMBER_OF_TOKEN_REGISTRY_ENTRIES = 1000

var LatestCheckedDeposits map[string]int64

func main() {

	var err error

	usr, err := user.Current()
	if err != nil {
		log.Error(err)
	}

	configFile := usr.HomeDir + "/.accumulatebridge/config.yaml"

	flag.StringVar(&configFile, "c", configFile, "config.yaml path")
	flag.Parse()

	LatestCheckedDeposits = make(map[string]int64)

	start(configFile)

}

func start(configFile string) {

	for {

		var err error
		var conf *config.Config
		var g *gnosis.Gnosis
		var e *evm.EVMClient
		var a *accumulate.AccumulateClient

		fmt.Println("Using config:", configFile)

		// init config
		if conf, err = config.NewConfig(configFile); err != nil {
			log.Fatal(err)
		}

		// set log level
		log.SetLevel(log.Lvl(conf.App.LogLevel))

		// init gnosis client
		if g, err = gnosis.NewGnosis(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Gnosis safe:", g.SafeAddress)
		fmt.Println("Bridge address:", g.BridgeAddress)
		fmt.Println("Gnosis API:", g.API)

		// init evm client
		if e, err = evm.NewEVMClient(conf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("EVM address:", e.PublicKey)
		fmt.Println("EVM API:", e.API)
		fmt.Println("EVM ChainId:", e.ChainId)

		// init gnosis client
		if a, err = accumulate.NewAccumulateClient(conf); err != nil {
			log.Fatal(err)
		}

		// init accumulate client
		fmt.Printf("Accumulate public key: %x\n", a.PublicKey)
		fmt.Printf("Accumulate public key hash: %x\n", a.PublicKeyHash)
		fmt.Println("Accumulate API:", a.API)
		fmt.Println("Bridge ADI:", a.ADI)

		// parse bridge fees
		bridgeFeesDataAccount := filepath.Join(conf.ACME.BridgeADI, accumulate.ACC_BRIDGE_FEES)
		fmt.Println("Getting bridge fees from", bridgeFeesDataAccount)
		fees, err := a.QueryLatestDataEntry(&accumulate.Params{URL: bridgeFeesDataAccount})
		if err != nil {
			fmt.Println("unable to get bridge fees from", bridgeFeesDataAccount)
			log.Fatal(err)
		}

		feesBytes, err := hex.DecodeString(fees.Data.Entry.Data[0])
		if err != nil {
			log.Error("can not decode entry data")
			log.Fatal(err)
		}

		err = json.Unmarshal(feesBytes, &global.BridgeFees)
		if err != nil {
			log.Error("unable to unmarshal entry data")
			log.Fatal(err)
		}

		// set chainId for tokens
		global.Tokens.ChainID = int64(conf.EVM.ChainId)

		fmt.Printf("Mint fee: %.2f%%\n", float64(global.BridgeFees.MintFee)/100)
		fmt.Printf("Burn fee: %.2f%%\n", float64(global.BridgeFees.BurnFee)/100)

		// parse token list from Accumulate
		// only once – when node is started
		// token list is mandatory, so return fatal error in case of error
		tokensDataAccount := filepath.Join(conf.ACME.BridgeADI, accumulate.ACC_TOKEN_REGISTRY)
		fmt.Println("Getting Accumulate tokens from", tokensDataAccount)
		tokens, err := a.QueryDataSet(&accumulate.Params{URL: tokensDataAccount, Count: int64(NUMBER_OF_TOKEN_REGISTRY_ENTRIES), Expand: true})
		if err != nil {
			fmt.Println("unable to get token list from", tokensDataAccount)
			log.Fatal(err)
		}

		fmt.Println("Got", len(tokens.Items), "data entry(s)")
		for _, item := range tokens.Items {
			parseToken(a, e, item)
		}

		fmt.Println("Found", len(global.Tokens.Items), "token(s)")

		if len(global.Tokens.Items) == 0 {
			log.Fatal("can not operate without tokens, shutting down")
		}

		// init interval go routines
		die := make(chan bool)
		go getStatus(a, die)
		go getLeader(a, die)
		// go debugLeader(die)

		go processBurnEvents(a, e, conf.EVM.BridgeAddress, die)
		go processNewDeposits(a, e, g, die)
		go submitEVMTxs(e, g, die)

		// init Accumulate Bridge API
		fmt.Println("Starting Accumulate Bridge API at port", conf.App.APIPort)
		log.Fatal(api.StartAPI(conf))

	}

}

// getLeader parses current leader's public key hash from Accumulate data account and compares it with Accumulate key in the config to find out if this node is a leader or not
func getLeader(a *accumulate.AccumulateClient, die chan bool) {

	leaderDataAccount := filepath.Join(a.ADI, accumulate.ACC_LEADER)

	for {

		select {
		default:

			leaderData, err := a.QueryLatestDataEntry(&accumulate.Params{URL: leaderDataAccount})
			if err != nil {
				fmt.Println("[leader] Unable to read bridge leader:", err)
				global.IsLeader = false
				global.IsAudit = false
				global.LeaderDuration = 0
			} else {
				fmt.Println("[leader] Bridge leader:", leaderData.Data.Entry.Data[0])
				decodedLeader, err := hex.DecodeString(leaderData.Data.Entry.Data[0])
				if err != nil {
					fmt.Println(err)
					global.IsLeader = false
					global.IsAudit = false
					global.LeaderDuration = 0
				}
				if bytes.Equal(decodedLeader, a.PublicKeyHash) {
					global.IsAudit = false
					global.LeaderDuration++
					if !global.IsLeader {
						if global.LeaderDuration <= LEADER_MIN_DURATION {
							fmt.Println("[leader] This node is leader, confirmations:", global.LeaderDuration, "of", LEADER_MIN_DURATION)
						}
						if global.LeaderDuration >= LEADER_MIN_DURATION {
							global.IsLeader = true
						}
					}
				} else {
					global.IsLeader = false
					global.IsAudit = true
					global.LeaderDuration = 0
				}
			}

			// check leader every minute
			time.Sleep(time.Duration(1) * time.Minute)

		case <-die:
			return
		}

	}

}

// getStatus checks if the bridge is online
func getStatus(a *accumulate.AccumulateClient, die chan bool) {

	statusDataAccount := filepath.Join(a.ADI, accumulate.ACC_BRIDGE_STATUS)

	for {

		select {
		default:

			online, err := a.QueryLatestDataEntry(&accumulate.Params{URL: statusDataAccount})
			if err != nil {
				fmt.Println("[status] Unable to read bridge status:", err)
				global.IsOnline = false
			} else {
				if len(online.Data.Entry.Data[0]) > 0 {
					fmt.Println("[status] Bridge is online")
					global.IsOnline = true
				} else {
					fmt.Println("[status] Bridge is paused")
					global.IsOnline = false
				}
			}

			// check leader every minute
			time.Sleep(time.Duration(1) * time.Minute)

		case <-die:
			return
		}

	}

}

// parseToken parses data entry with token information received from data account
func parseToken(a *accumulate.AccumulateClient, e *evm.EVMClient, entry *accumulate.DataEntry) {

	fmt.Println("Parsing", entry.EntryHash)

	tokenEntry := &schema.TokenEntry{}

	// check version
	if len(entry.Entry.Data) < 2 {
		log.Debug("looking for at least 2 data fields in entry, found ", len(entry.Entry.Data))
		return
	}

	version, err := hex.DecodeString(entry.Entry.Data[0])
	if err != nil {
		log.Debug("can not decode entry data")
		return
	}

	if !bytes.Equal(version, []byte(accumulate.TOKEN_REGISTRY_VERSION)) {
		log.Debug("entry version is not ", accumulate.TOKEN_REGISTRY_VERSION)
		return
	}

	// convert entry data to bytes
	tokenData, err := hex.DecodeString(entry.Entry.Data[1])
	if err != nil {
		log.Debug("can not decode entry data")
		return
	}

	// try to unmarshal the entry
	err = json.Unmarshal(tokenData, tokenEntry)
	if err != nil {
		log.Debug("unable to unmarshal entry data")
		return
	}

	// if entry is disabled, skip
	if !tokenEntry.Enabled {
		log.Debug("token is disabled")
		return
	}

	// validate token
	validate := validator.New()
	err = validate.Struct(tokenEntry)
	if err != nil {
		log.Debug(err)
		return
	}

	token := &schema.Token{}

	for _, wrappedToken := range tokenEntry.Wrapped {
		// search for current chainid
		if wrappedToken.ChainID == global.Tokens.ChainID {
			err = validate.Struct(wrappedToken)
			if err != nil {
				log.Debug(err)
				return
			}
			token.EVMAddress = wrappedToken.Address
			token.EVMMintTxCost = wrappedToken.MintTxCost
		}
	}

	// if no token address found, error
	if token.EVMAddress == "" {
		log.Debug("can not find token address for chainid ", global.Tokens.ChainID)
		return
	}

	// parse token info from Accumulate
	t, err := a.QueryToken(&accumulate.Params{URL: tokenEntry.URL})
	if err != nil {
		log.Debug("can not get token from accumulate api ", err)
		return
	}

	token.URL = t.Data.URL
	token.Symbol = t.Data.Symbol
	token.Precision = t.Data.Precision

	// check if bridge has token account on this chain for this token
	tokenAccountUrl := accumulate.GenerateTokenAccount(a.ADI, global.Tokens.ChainID, token.Symbol)
	_, err = a.QueryTokenAccount(&accumulate.Params{URL: tokenAccountUrl})
	if err != nil {
		log.Debug("can not get token account ", tokenAccountUrl, " from accumulate api")
		return
	}

	// parse token info from Ethereum
	evmT, err := e.GetERC20(token.EVMAddress)
	if err != nil {
		log.Debug("can not get token from ethereum api ", err)
	}
	token.EVMSymbol = evmT.Symbol
	token.EVMDecimals = evmT.Decimals

	// check for duplicates, if found override
	exists := utils.SearchAccumulateToken(token.URL)

	// if not found, append new token
	if exists == nil {
		log.Info("added token ", token.URL)
		global.Tokens.Items = append(global.Tokens.Items, token)
		return
	}

	log.Info("duplicate token ", token.URL, ", overwritten")
	*exists = *token

}

// debugLeader helps to debug leader behaviour
func debugLeader(die chan bool) {

	for {

		select {
		default:

			log.Debug("isLeader=", global.IsLeader)
			time.Sleep(time.Duration(5) * time.Second)

		case <-die:
			return
		}

	}

}

// processBurnEvents
func processBurnEvents(a *accumulate.AccumulateClient, e *evm.EVMClient, bridge string, die chan bool) {

	for {

		select {
		default:

			time.Sleep(time.Duration(30) * time.Second)

			releaseQueue := accumulate.GenerateReleaseDataAccount(a.ADI, int64(e.ChainId), accumulate.ACC_RELEASE_QUEUE)

			if global.IsOnline {

				if global.IsLeader {

					fmt.Println("[release] Checking pending chain of", releaseQueue)

					pendingEntries, err := a.QueryPendingChain(&accumulate.Params{URL: releaseQueue})
					if err != nil {
						fmt.Println("[release] Stopping the process, unable to get pending chain:", err)
						break
					}

					// if there are any pending entries, do not produce new tx
					if len(pendingEntries.Items) > 0 {
						fmt.Println("[release] Stopping the process, found pending entries in", releaseQueue)
						break
					}

					fmt.Println("[release] Getting block height from the latest entry of", releaseQueue)
					latestReleaseEntry, err := a.QueryLatestDataEntry(&accumulate.Params{URL: releaseQueue})

					// if Accumulate does not return blockheight, shut down to prevent double spending
					if err != nil {
						fmt.Println("[release] Unable to get block height:", err)
						break
					}

					// parse latest burn entry to find out evm blockHeight
					burnEntry, err := schema.ParseBurnEvent(latestReleaseEntry.Data)
					if err != nil {
						fmt.Println("[release]", err)
						break
					}

					// looking for evm logs starting from latest height+1
					start := burnEntry.BlockHeight + 1

					fmt.Println("[release] Parsing new EVM events for", bridge, "starting from blockHeight", start)
					logs, err := e.ParseBridgeLogs("Burn", bridge, start)
					if err != nil {
						fmt.Println("[release]", err)
						break
					}

					knownHeight := 0

					// logs are sorted by timestamp asc
					for _, l := range logs {

						fmt.Println("[release] Height", l.BlockHeight, "txid", l.TxID.Hex())

						// additional check in case evm node returns invalid response
						if int64(l.BlockHeight) < start {
							fmt.Println("[release] Invalid height, expected height >=", start)
							continue
						}

						// process only single block height at once
						// if blockheight changed = shutdown
						if knownHeight > 0 && l.BlockHeight != uint64(knownHeight) {
							fmt.Println("[release] Height changed, will process event in the next batch, stopping the process")
							break
						}

						// create burnEntry
						burnEntry := &schema.BurnEvent{}
						burnEntry.EVMTxID = l.TxID.Hex()
						burnEntry.BlockHeight = int64(l.BlockHeight)
						burnEntry.TokenAddress = l.Token.String()
						burnEntry.Destination = l.Destination
						burnEntry.Amount = l.Amount.Int64()

						// find token
						token := utils.SearchEVMToken(burnEntry.TokenAddress)

						// skip if no token found
						if token == nil {
							continue
						}

						fmt.Println("[release] Sending", burnEntry.Amount, token.Symbol, "to", burnEntry.Destination)

						// generate accumulate token tx
						txhash, err := a.SendTokens(burnEntry.Destination, burnEntry.Amount, token.URL, int64(e.ChainId))
						if err != nil {
							fmt.Println("[release] tx failed:", err)
							continue
						}

						fmt.Println("[release] tx sent:", txhash)

						burnEntry.TxHash = txhash

						burnEntryBytes, err := json.Marshal(burnEntry)
						if err != nil {
							fmt.Println("[release] can not marshal burn entry:", err)
							continue
						}

						var content [][]byte
						content = append(content, []byte(accumulate.RELEASE_QUEUE_VERSION))
						content = append(content, burnEntryBytes)

						entryhash, err := a.WriteData(releaseQueue, content)
						if err != nil {
							fmt.Println("[release] data entry creation failed:", err)
							continue
						}

						fmt.Println("[release] data entry created:", entryhash)

						knownHeight = int(l.BlockHeight)

					}

				} else if global.IsAudit {

					fmt.Println("[release] Checking pending chain of", releaseQueue)

					pending, err := a.QueryPendingChain(&accumulate.Params{URL: releaseQueue})
					if err != nil {
						fmt.Println("[release] can not get pending data entries:", err)
						break
					}

					// if no pending entries, shut down
					if len(pending.Items) == 0 {
						fmt.Println("[release] Stopping the process, no pending entries found in", releaseQueue)
						break
					}

					fmt.Println("[release] Getting block height from the latest entry of", releaseQueue)
					latestReleaseEntry, err := a.QueryLatestDataEntry(&accumulate.Params{URL: releaseQueue})

					// if Accumulate does not return blockheight, shut down to prevent double spending
					if err != nil {
						fmt.Println("[release] Unable to get block height:", err)
						break
					}

					// parse latest burn entry to find out evm blockHeight
					latestCompletedBurn, err := schema.ParseBurnEvent(latestReleaseEntry.Data)
					if err != nil {
						fmt.Println("[release]", err)
						break
					}

					// looking for pending tx with blockheight starting from latest height+1
					start := latestCompletedBurn.BlockHeight + 1

					for _, entryhash := range pending.Items {

						fmt.Println("[release] processing pending entry", entryhash)

						entryURL := entryhash + "@" + releaseQueue
						entry, err := a.QueryDataEntry(&accumulate.Params{URL: entryURL})
						if err != nil {
							fmt.Println("[release] Unable to get data entry", err)
							continue
						}

						burnEntry, err := schema.ParseBurnEvent(entry.Data)
						if err != nil {
							fmt.Println("[release] Unable to parse burn event from data entry", err)
							continue
						}

						fmt.Println("[release] start", start, "event blockheight", burnEntry.BlockHeight)

						// check block height to avoid old txs
						if int64(burnEntry.BlockHeight) < start {
							fmt.Println("[release] Invalid height, expected height >=", start)
							continue
						}

						// find token
						token := utils.SearchEVMToken(burnEntry.TokenAddress)

						// skip if no token found
						if token == nil {
							continue
						}

						fmt.Println("[release] Found new pending tx:", burnEntry.TxHash, "- Sending", burnEntry.Amount, token.Symbol, "to", burnEntry.Destination)
						fmt.Println("[release] Checking corresponding EVM tx:", burnEntry.EVMTxID)

						// parse evm tx
						evmTx, err := e.GetTx(burnEntry.EVMTxID)
						if err != nil {
							fmt.Println(err)
							continue
						}

						burnData, err := abiutil.UnpackBurnTxInputData(evmTx.Data)
						if err != nil {
							fmt.Println("[release] unable to read burn tx:", err)
							continue
						}

						// validate burn entry against evm tx
						err = utils.ValidateBurnEntry(burnEntry, burnData)
						if err != nil {
							fmt.Println("[release] burn entry validation failed:", err)
							continue
						}

						// parse accumulate txid
						txid, err := acmeurl.ParseTxID(burnEntry.TxHash)
						if err != nil {
							fmt.Println(err)
							continue
						}

						remoteTxHash := txid.Hash()

						// parse accumulate tx
						tx, err := a.QueryTokenTx(&accumulate.Params{URL: burnEntry.TxHash})
						if err != nil {
							fmt.Println(err)
							continue
						}

						// validate accumulate tx against evm tx
						err = utils.ValidateReleaseTx(tx.Data, burnData)
						if err != nil {
							fmt.Println("[release] accumulate tx validation failed:", err)
							continue
						}

						// sign accumulate tx
						txhash, err := a.RemoteTransaction(hex.EncodeToString(remoteTxHash[:]))
						if err != nil {
							fmt.Println("[release] tx failed:", err)
							continue
						}

						fmt.Println("[release] tx sent:", txhash)

						// sign data entry
						txhash, err = a.RemoteTransaction(entryhash)
						if err != nil {
							fmt.Println("[release] tx failed:", err)
							continue
						}

						fmt.Println("[release] tx sent:", txhash)

					}
				}

			}

		case <-die:
			return
		}

	}

}

// processNewDeposits
func processNewDeposits(a *accumulate.AccumulateClient, e *evm.EVMClient, g *gnosis.Gnosis, die chan bool) {

	for {

		select {
		default:

			time.Sleep(time.Duration(30) * time.Second)

			if global.IsOnline {

				if global.IsLeader {

					for _, token := range global.Tokens.Items {

						// get gnosis safe
						safe, err := g.GetSafe()
						if err != nil {
							fmt.Println("[mint] can not get gnosis safe:", err)
							break
						}

						// check if there are pending txs at current nonce
						safeTxs, err := g.GetSafeMultisigTxs()
						if err != nil {
							fmt.Println("[mint] can not get gnosis safe multisig txs:", err)
							break
						}

						if len(safeTxs.Results) > 0 {
							if safeTxs.Results[0].Nonce >= safe.Nonce {
								fmt.Println("[mint] stopping the process, gnosis safe has unprocessed tx with nonce", safeTxs.Results[0].Nonce)
								break
							}
						}

						mintQueue := accumulate.GenerateMintDataAccount(a.ADI, int64(e.ChainId), accumulate.ACC_MINT_QUEUE, token.Symbol)

						fmt.Println("[mint] Checking pending chain of", mintQueue)

						pendingEntries, err := a.QueryPendingChain(&accumulate.Params{URL: mintQueue})
						if err != nil {
							fmt.Println("[mint] Stopping the process, unable to get pending chain:", err)
							continue
						}

						// if there are any pending entries, do not produce new tx
						if len(pendingEntries.Items) > 0 {
							fmt.Println("[mint] Stopping the process, found pending entries in", mintQueue)
							continue
						}

						fmt.Println("[mint] Getting seq number from the latest entry of", mintQueue)
						latestMintEntry, err := a.QueryLatestDataEntry(&accumulate.Params{URL: mintQueue})

						// if Accumulate does not return seq number, shut down to prevent double minting
						if err != nil {
							fmt.Println("[mint] Unable to get seq number:", err)
							continue
						}

						// parse latest mint entry to find out seq number
						mintEntry, err := schema.ParseDepositEvent(latestMintEntry.Data)
						if err != nil {
							fmt.Println("[mint] Unable to parse deposit event from data entry", err)
							continue
						}

						// looking for accumulate token txs starting from latest height+1
						start := mintEntry.SeqNumber + 1
						if LatestCheckedDeposits[token.Symbol] > mintEntry.SeqNumber {
							start = LatestCheckedDeposits[token.Symbol] + 1
						}

						tokenAccount := accumulate.GenerateTokenAccount(a.ADI, int64(e.ChainId), token.Symbol)

						fmt.Println("[mint] Parsing new accumulate token txs in", tokenAccount, "starting from seq number", start)

						count := int64(NUMBER_OF_ACCUMULATE_TOKEN_TXS)

						txs, err := a.QueryTxHistory(&accumulate.Params{URL: tokenAccount, Start: start, Count: count})
						if err != nil {
							fmt.Println("[mint] Unable to get tx history for", tokenAccount, err)
							continue
						}

						fmt.Println("[mint] Found", len(txs.Items), "txs in", tokenAccount, "seq number from", start, "to", start+count)

						// cursor is to track seq number and update it in map in the end
						cursor := start

						if len(txs.Items) > 0 {

							for i, tx := range txs.Items {

								// cursor = current seq number
								cursor = int64(i) + start

								// validate tx
								fmt.Println("[mint] Validating tx", tx.TxHash, "seq number", cursor)
								err := utils.ValidateDepositTx(tx)
								if err != nil {
									fmt.Println("[mint] tx validation failed:", err)
									continue
								}

								// query cause tx
								cause, err := a.QueryTokenTx(&accumulate.Params{URL: tx.Data.Cause})
								if err != nil {
									fmt.Println("[mint] can not get cause tx:", err)
									// if we are here, then something happened on the accumulate api side
									// reset cursor and break to start over
									cursor = start - 1
									break
								}

								// validate cause tx
								err = utils.ValidateCauseTx(cause)
								if err != nil {
									fmt.Println("[mint] cause tx validation failed:", err)
									continue
								}

								amount := new(big.Int)
								amount, ok := amount.SetString(tx.Data.Amount, 10)
								if !ok {
									fmt.Println("[mint] unable to convert tx amount")
									// if we are here, then something unexpected happened
									// reset cursor and break to start over
									cursor = start - 1
									break
								}

								// create mintEntry
								mintEntry := &schema.DepositEvent{}
								mintEntry.Amount = amount.Int64()
								mintEntry.Destination = cause.Transaction.Header.Memo
								mintEntry.SeqNumber = cursor
								mintEntry.Source = cause.Data.From
								mintEntry.TokenAddress = token.EVMAddress
								mintEntry.TokenURL = token.URL
								mintEntry.TxID = tx.TxID

								// generate mint tx data
								data, err := abiutil.GenerateMintTxData(token.EVMAddress, cause.Transaction.Header.Memo, amount)
								if err != nil {
									fmt.Println("[mint] can not generate mint tx:", err)
									// if we are here, then something unexpected happened
									// reset cursor and break to start over
									cursor = start - 1
									break
								}

								// generate gnosis safe tx
								contractHash, signature, err := g.SignMintTx(token.EVMAddress, cause.Transaction.Header.Memo, amount)
								if err != nil {
									fmt.Println("[mint] can not sign mint tx:", err)
									// if we are here, then something unexpected happened
									// reset cursor and break to start over
									cursor = start - 1
									break
								}

								// submit multisig tx to the gnosis safe api
								safeTx := gnosis.NewMultisigTx{}
								safeTx.To = g.BridgeAddress
								safeTx.Data = hexutil.Encode(data)
								safeTx.GasToken = abiutil.ZERO_ADDR
								safeTx.RefundReceiver = abiutil.ZERO_ADDR
								safeTx.Nonce = safe.Nonce
								safeTx.ContractTransactionHash = hexutil.Encode(contractHash)
								safeTx.Sender = g.PublicKey.Hex()
								safeTx.Signature = hexutil.Encode(signature)

								err = g.CreateSafeMultisigTx(&safeTx)
								if err != nil {
									fmt.Println("[mint] gnosis safe api error:", err)
									// if we are here, then something happened on the gnosis api side
									// reset cursor and break to start over
									cursor = start - 1
									break
								}

								// create accumulate data entry
								mintEntry.SafeTxHash = hexutil.Encode(contractHash)
								mintEntry.SafeTxNonce = safe.Nonce

								mintEntryBytes, err := json.Marshal(mintEntry)
								if err != nil {
									fmt.Println("[mint] can not marshal mint entry:", err)
									// if we are here, then something unexpected happened
									// reset cursor and break to start over
									cursor = start - 1
									break
								}

								var content [][]byte
								content = append(content, []byte(accumulate.MINT_QUEUE_VERSION))
								content = append(content, mintEntryBytes)

								entryhash, err := a.WriteData(mintQueue, content)
								if err != nil {
									fmt.Println("[mint] data entry creation failed:", err)
									// if we are here, then something happened on the accumulate api side
									// reset cursor and break to start over
									cursor = start - 1
									break
								}

								fmt.Println("[mint] data entry created:", entryhash)
								break

							}

						} else {

							// if no tx found, move cursor back by 1 to start over from the same seq number
							cursor--

						}

						LatestCheckedDeposits[token.Symbol] = cursor

					}

				} else if global.IsAudit {

					for _, token := range global.Tokens.Items {

						// get gnosis safe
						safe, err := g.GetSafe()
						if err != nil {
							fmt.Println("[mint] can not get gnosis safe:", err)
							break
						}

						mintQueue := accumulate.GenerateMintDataAccount(a.ADI, int64(e.ChainId), accumulate.ACC_MINT_QUEUE, token.Symbol)
						depositTokenAccount := accumulate.GenerateTokenAccount(a.ADI, int64(e.ChainId), token.Symbol)

						fmt.Println("[mint] Checking pending chain of", mintQueue)

						pending, err := a.QueryPendingChain(&accumulate.Params{URL: mintQueue})
						if err != nil {
							fmt.Println("[mint] can not get pending data entries:", err)
							continue
						}

						// if no pending entries, shut down
						if len(pending.Items) == 0 {
							fmt.Println("[mint] Stopping the process, no pending entries found in", mintQueue)
							continue
						}

						fmt.Println("[mint] Getting sequence number from the latest entry of", mintQueue)
						latestReleaseEntry, err := a.QueryLatestDataEntry(&accumulate.Params{URL: mintQueue})

						// if Accumulate does not return sequence number, shut down to prevent double minting
						if err != nil {
							fmt.Println("[mint] Unable to get sequence number:", err)
							break
						}

						// parse latest mint entry to find out sequence number
						latestCompletedMint, err := schema.ParseDepositEvent(latestReleaseEntry.Data)
						if err != nil {
							fmt.Println("[mint]", err)
							break
						}

						// looking for pending tx with sequence number starting from latest seq number+1
						start := latestCompletedMint.SeqNumber + 1

						for _, entryhash := range pending.Items {

							fmt.Println("[mint] processing pending entry", entryhash)

							entryURL := entryhash + "@" + mintQueue
							entry, err := a.QueryDataEntry(&accumulate.Params{URL: entryURL})
							if err != nil {
								fmt.Println("[mint] Unable to get data entry", err)
								continue
							}

							mintEntry, err := schema.ParseDepositEvent(entry.Data)
							if err != nil {
								fmt.Println("[mint] Unable to parse deposit event from data entry", err)
								continue
							}

							fmt.Println("[mint] start", start, "event seq number", mintEntry.SeqNumber)

							// check block height to avoid old txs
							if int64(mintEntry.SeqNumber) < start {
								fmt.Println("[mint] Invalid seq number, expected seq number >=", start)
								continue
							}

							fmt.Println("[mint] Found new pending tx:", mintEntry.TxID, "- Minting", mintEntry.Amount, token.EVMSymbol, "to", mintEntry.Destination)
							fmt.Println("[mint] Checking corresponding Accumulate tx seq number:", mintEntry.SeqNumber)

							// parse tx using seq number
							txs, err := a.QueryTxHistory(&accumulate.Params{URL: depositTokenAccount, Count: 1, Start: mintEntry.SeqNumber})
							if err != nil {
								fmt.Println(err)
								continue
							}

							// if tx found
							if len(txs.Items) > 0 {
								fmt.Println("[mint] tx with seq number:", txs.Items[0].TxID)
								fmt.Println("[mint] mint entry tx:", mintEntry.TxID)
								// validate txid in mint entry
								if txs.Items[0].TxID != mintEntry.TxID {
									continue
								}
							} else {
								fmt.Println("[mint] not found tx by seq number:", mintEntry.SeqNumber)
								continue
							}

							// query cause tx
							cause, err := a.QueryTokenTx(&accumulate.Params{URL: txs.Items[0].Data.Cause})
							if err != nil {
								fmt.Println("[mint] can not get cause tx:", err)
								continue
							}

							// validate cause tx
							err = utils.ValidateCauseTx(cause)
							if err != nil {
								fmt.Println("[mint] cause tx validation failed:", err)
								continue
							}

							// validate mint entry against accumulate txs
							err = utils.ValidateMintEntry(mintEntry, txs.Items[0], cause)
							if err != nil {
								fmt.Println("[mint] accumulate tx validation failed:", err)
								continue
							}

							// check mint entry safe tx nonce
							if mintEntry.SafeTxNonce != safe.Nonce {
								fmt.Println("[mint] mint entry safe tx nonce:", mintEntry.SafeTxNonce, "safe nonce:", safe.Nonce)
								continue
							}

							fmt.Println("[mint] Generating and signing gnosis safe tx")

							// generate mint tx data
							amount := big.NewInt(mintEntry.Amount)
							data, err := abiutil.GenerateMintTxData(token.EVMAddress, cause.Transaction.Header.Memo, amount)
							if err != nil {
								fmt.Println("[mint] can not generate mint tx:", err)
								continue
							}

							// generate gnosis safe tx
							contractHash, signature, err := g.SignMintTx(token.EVMAddress, cause.Transaction.Header.Memo, amount)
							if err != nil {
								fmt.Println("[mint] can not sign mint tx:", err)
								continue
							}

							// check if contract hash == mint entry safetxhash
							if hexutil.Encode(contractHash) != mintEntry.SafeTxHash {
								fmt.Println("[mint] mint entry safe tx hash:", mintEntry.SafeTxHash, "generated safe tx hash:", contractHash)
								continue
							}

							// submit multisig tx to the gnosis safe api
							safeTx := gnosis.NewMultisigTx{}
							safeTx.To = g.BridgeAddress
							safeTx.Data = hexutil.Encode(data)
							safeTx.GasToken = abiutil.ZERO_ADDR
							safeTx.RefundReceiver = abiutil.ZERO_ADDR
							safeTx.Nonce = safe.Nonce
							safeTx.ContractTransactionHash = hexutil.Encode(contractHash)
							safeTx.Sender = g.PublicKey.Hex()
							safeTx.Signature = hexutil.Encode(signature)

							err = g.CreateSafeMultisigTx(&safeTx)
							if err != nil {
								fmt.Println("[mint] gnosis safe api error:", err)
								continue
							}

							fmt.Println("[mint] gnosis safe tx signed: nonce", mintEntry.SafeTxNonce, "safeTxHash", mintEntry.SafeTxHash)

							// sign data entry
							txhash, err := a.RemoteTransaction(entryhash)
							if err != nil {
								fmt.Println("[mint] tx failed:", err)
								continue
							}

							fmt.Println("[mint] tx sent:", txhash)

						}

					}
				}
			}

		case <-die:
			return
		}

	}

}

// submitEVMTxs
func submitEVMTxs(e *evm.EVMClient, g *gnosis.Gnosis, die chan bool) {

	for {

		select {
		default:

			time.Sleep(time.Duration(30) * time.Second)

			if global.IsOnline {

				if global.IsLeader {

					// get gnosis safe
					safe, err := g.GetSafe()
					if err != nil {
						fmt.Println("[submit] can not get gnosis safe:", err)
						break
					}

					txs, err := g.GetSafeMultisigTxByNonce(safe.Nonce)
					if err != nil {
						fmt.Println("[submit] can not get gnosis safe txs:", err)
						break
					}

					for _, tx := range txs.Results {

						fmt.Println("[submit] found safetxhash:", tx.SafeTxHash)

						// check number of signatures
						if len(tx.Confirmations) < int(safe.Threshold) {
							fmt.Println("[submit]", len(tx.Confirmations), "signatures,", safe.Threshold, "required")
							break
						}

						// sort signatures
						sort.Slice(tx.Confirmations, func(i, j int) bool {
							switch strings.Compare(tx.Confirmations[i].Owner, tx.Confirmations[j].Owner) {
							case -1:
								return true
							case 1:
								return false
							}
							return tx.Confirmations[i].Owner > tx.Confirmations[j].Owner
						})

						// concatenate signatures
						var sig []byte
						for _, con := range tx.Confirmations {
							sigBytes, err := hexutil.Decode(con.Signature)
							if err != nil {
								fmt.Println("[submit] can not decode signature hex:", err)
								break
							}
							sig = append(sig, sigBytes...)
						}

						// generate tx input data
						txData, err := abiutil.GenerateExecTransaction(g.BridgeAddress, tx.Data, hexutil.Encode(sig))
						if err != nil {
							fmt.Println("[submit] can not generate tx data:", err)
							break
						}

						to := common.HexToAddress(g.SafeAddress)

						// submit ethereum tx
						sentTx, err := e.SubmitEIP1559Tx(gnosis.MINT_GAS_LIMIT, e.MaxGasFee, e.MaxPriorityFee, &to, 0, txData)
						if err != nil {
							fmt.Println("[submit] ethereum tx error:", err)
							break
						}

						fmt.Println("[submit] tx sent:", sentTx.Hash().Hex())

					}

				}
			}

		case <-die:
			return
		}

	}

}
