package cpmm

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("cpmm",
		fx.Provide(
			NewClient,
		),
	)
}
