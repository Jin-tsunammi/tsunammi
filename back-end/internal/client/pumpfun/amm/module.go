package pump_amm

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("pump_amm",
		fx.Provide(
			NewClient,
		),
	)
}
