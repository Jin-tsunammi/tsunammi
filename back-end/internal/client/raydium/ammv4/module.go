package ammv4

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("ammv4",
		fx.Provide(
			NewClient,
		),
	)
}
