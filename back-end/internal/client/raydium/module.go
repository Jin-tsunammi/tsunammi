package raydium

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("raydium",
		fx.Provide(
			NewClient,
		),
	)
}
