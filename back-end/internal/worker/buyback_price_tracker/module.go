package buybackpricetracker

import (
	"context"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("worker-buyback-price-tracker",
		fx.Provide(
			NewBuybackPriceTracker,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, worker *BuybackPriceTracker) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return worker.Start(ctx)
					},
					OnStop: func(ctx context.Context) error {
						return worker.Stop(ctx)
					},
				})
			},
		),
	)
}
