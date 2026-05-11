package lighthouse

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("lighthouse_client",
		fx.Provide(NewClient),
	)
}
