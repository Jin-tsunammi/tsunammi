package pool

import (
	"context"
	"errors"
	"mm/config"

	"github.com/weeaa/jito-go/clients/searcher_client"
	"go.uber.org/fx"
)

var Module = fx.Module("pool",
	fx.Provide(
		func(data []*searcher_client.Client, cfg *config.Config) *CloseableRoundRobin[*searcher_client.Client] {
			return NewCloseableRoundRobin(data, cfg.Jito.CoolDown)
		},
	),
	fx.Invoke(
		func(lc fx.Lifecycle, pool *CloseableRoundRobin[*searcher_client.Client]) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return pool.Start(ctx)
				},
				OnStop: func(ctx context.Context) error {
					errC := pool.Close()
					errS := pool.Stop(ctx)
					return errors.Join(errC, errS)
				},
			})
		},
	),
)
