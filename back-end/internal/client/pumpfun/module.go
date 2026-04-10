package pumpfun

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("pumpfun",
		fx.Provide(
			NewClient,
		),
	)
}
