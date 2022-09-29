package api

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/AccumulateNetwork/bridge/accumulate"
	"github.com/AccumulateNetwork/bridge/config"
	"github.com/AccumulateNetwork/bridge/global"
	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/transport"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	r      *rpc.RpcServer
	a      *accumulate.AccumulateClient
}

func StartAPI(conf *config.Config) error {

	a, err := accumulate.NewAccumulateClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	r := rpc.New(
		rpc.WithTransport(&transport.HTTP{Bind: ":" + strconv.Itoa(conf.App.APIPort), CORSOrigin: "*"}), // HTTP transport
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	s := &Server{
		ctx:    ctx,
		cancel: cancel,
		r:      r,
		a:      a,
	}

	s.r.Register("fees", rpc.H(s.Fees))
	s.r.Register("tokens", rpc.H(s.Tokens))
	s.r.Register("token-account", rpc.H(s.TokenAccount))

	if err := s.r.Run(ctx); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *Server) Fees(ctx context.Context, _ *NoArgs) (interface{}, error) {
	return &global.BridgeFees, nil
}

func (s *Server) Tokens(ctx context.Context, _ *NoArgs) (interface{}, error) {
	return &global.Tokens, nil
}

func (s *Server) TokenAccount(ctx context.Context, url *URL) (interface{}, error) {

	account, err := s.a.QueryTokenAccount(&accumulate.Params{URL: url.URL})
	if err != nil {
		return nil, err
	}

	return account.Data, nil

}

type NoArgs struct {
}

type URL struct {
	URL string `json:"url"`
}
