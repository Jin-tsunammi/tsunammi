package kucoinapi

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("kucoinapi",
		fx.Provide(NewKuCoinApiClient),
	)
}
