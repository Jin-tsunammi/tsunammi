package cache

import (
	"context"
	"mm/config"

	"github.com/gagliardetto/solana-go"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("rate_cache",

		fx.Provide(
			fx.Annotate(
				NewRateStorage,
				fx.As(
					new(RateStorage),
				)),
		),
		fx.Provide(
			fx.Annotate(
				NewJitoValidatorStorage,
				fx.As(
					new(JitoValidatorStorage),
				)),
		),

		fx.Provide(
			fx.Annotate(
				NewJitoScheduleStorage,
				fx.As(
					new(JitoScheduleStorage),
				)),
		),

		fx.Provide(
			fx.Annotate(
				NewProjectStorage,
				fx.As(
					new(ProjectStorage),
				)),
		),

		fx.Invoke(
			func(lc fx.Lifecycle, cfg *config.Config, cache RateStorage) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						rateStore := cache.(*rateStorage)
						err := rateStore.jupiterUpdateRateAndDecimals(
							ctx,
							solana.SolMint,
						)
						if err != nil {
							return err
						}
						return nil
					},
				},
				)
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, cfg *config.Config, cache JitoValidatorStorage) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						rateStore := cache.(*jitoValidatorStorage)
						return rateStore.update(
							ctx,
							cfg.App.ExchangeRateCacheTTL,
						)
					},
				},
				)
			},
		),

		fx.Invoke(
			func(lc fx.Lifecycle, cfg *config.Config, cache JitoScheduleStorage) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						rateStore := cache.(*jitoScheduleStorage)
						return rateStore.update(
							ctx,
							cfg.Jito.ScheduleTTL,
						)
					},
				},
				)
			},
		),
	)
}
