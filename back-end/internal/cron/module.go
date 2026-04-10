package cron

import (
	rcron "github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("cron",
		fx.Provide(rcron.New),
		fx.Provide(NewDBCron),
		fx.Invoke(
			func(lc fx.Lifecycle, cron *DBCron) {
				lc.Append(fx.Hook{
					OnStart: cron.start,
					OnStop:  cron.stop,
				})
			},
		),
	)
}
