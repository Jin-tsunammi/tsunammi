package kucoinapi

import (
	"go.uber.org/fx"
	"resty.dev/v3"
)

func Module() fx.Option {
	return fx.Module("kucoinapi",
		fx.Provide(resty.New),
		fx.Provide(NewKuCoinApiClient),
	)
}
