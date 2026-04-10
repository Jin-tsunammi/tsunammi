package bonding

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("pump_bonding",
		fx.Provide(
			NewClient,
		),
	)
}
