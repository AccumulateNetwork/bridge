package api

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"
	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/transport"
)

func StartAPI(conf *config.Config) error {

	s := rpc.New(
		rpc.WithTransport(&transport.HTTP{Bind: ":" + strconv.Itoa(conf.App.APIPort), CORSOrigin: "*"}), // HTTP transport
	)

	s.Register("fees", rpc.H(Fees))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}

	return nil
}

func Fees(ctx context.Context, _ *NoArgs) (*BridgeFees, error) {
	return &BridgeFees{}, nil
}

type BridgeFees struct {
	MintFee int64 `json:"mintFee"`
	EVMFee  int64 `json:"evmFee"`
	BurnFee int64 `json:"burnFee"`
}

type NoArgs struct {
}
