package api

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/global"
	"github.com/AccumulateNetwork/bridge/schema"
	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/transport"
)

func StartAPI(conf *config.Config) error {

	s := rpc.New(
		rpc.WithTransport(&transport.HTTP{Bind: ":" + strconv.Itoa(conf.App.APIPort), CORSOrigin: "*"}), // HTTP transport
	)

	s.Register("fees", rpc.H(Fees))
	s.Register("tokens", rpc.H(Tokens))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}

	return nil
}

func Fees(ctx context.Context, _ *NoArgs) (*schema.BridgeFees, error) {
	return &global.BridgeFees, nil
}

func Tokens(ctx context.Context, _ *NoArgs) (*schema.Tokens, error) {
	return &global.Tokens, nil
}

type NoArgs struct {
}
