package solanarpc

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("solana_rpc",
		fx.Provide(
			fx.Annotate(
				NewSolanaRPCClient,
				fx.As(
					new(SolanaRPC),
				),
			),
		),
	)
}
