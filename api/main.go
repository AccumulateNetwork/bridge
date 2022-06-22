package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/rpc/middleware"
	"go.neonxp.dev/jsonrpc2/transport"
)

func main() {

	s := rpc.New(
		rpc.WithLogger(rpc.StdLogger),                                      // Optional logger
		rpc.WithTransport(&transport.HTTP{Bind: ":8000", CORSOrigin: "*"}), // HTTP transport
	)

	s.Use(
		rpc.WithTransport(&transport.TCP{Bind: ":3000"}),     // TCP transport
		rpc.WithMiddleware(middleware.Logger(rpc.StdLogger)), // Logger middleware
	)
	s.Register("bridgeFee", rpc.H(BridgeFee))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func BridgeFee(ctx context.Context, args *Args) (*Fee, error) {
	return &Fee{BridgeFee: "1111"}, nil
}

type Fee struct {
	BridgeFee string `json:"bridgeFee"`
}

type Args struct {
}
