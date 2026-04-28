package buyback

import (
	"context"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("worker-buyback-campaign-manager",
		fx.Provide(
			NewSmartBuybackManager,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, manager *CampaignManager) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return manager.Start(ctx)
					},
					OnStop: func(ctx context.Context) error {
						return manager.Stop(ctx)
					},
				})
			},
		),
	)
}
