package worker

import (
	"context"
	"mm/config"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("worker",
		fx.Provide(
			NewSwapTargetManager,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, manager *SwapTargetManager, cfg *config.Config) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go manager.controlThread(cfg)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						manager.close()
						return nil
					},
				})
			},
		),
	)
}
