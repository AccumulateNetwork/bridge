package api

import (
	"context"

	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/rpc/middleware"
	"go.neonxp.dev/jsonrpc2/transport"
)

func api() {
	s := rpc.New(
		rpc.WithLogger(rpc.StdLogger),                     // Optional logger
		rpc.WithTransport(&transport.HTTP{Bind: ":8000"}), // HTTP transport
	)

	s.Use(
		rpc.WithTransport(&transport.TCP{Bind: ":3000"}),     // TCP transport
		rpc.WithMiddleware(middleware.Logger(rpc.StdLogger)), // Logger middleware
	)
	s.Register("multiply", rpc.H(Multiply))
	s.Run(context.Background())
}

func Multiply(ctx context.Context, args *Args) (int, error) {
	return args.A * args.B, nil
}

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Quotient struct {
	Quo int `json:"quo"`
	Rem int `json:"rem"`
}
