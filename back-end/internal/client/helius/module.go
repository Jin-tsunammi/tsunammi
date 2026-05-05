package helius

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("helius_client",
		fx.Provide(NewClient),
	)
}
