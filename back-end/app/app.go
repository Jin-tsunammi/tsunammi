package app

import (
	"mm/config"
	"mm/internal/client/jito"
	"mm/internal/client/kucoinapi"
	"mm/internal/client/pumpfun"
	pumpAMM "mm/internal/client/pumpfun/amm"
	pumpBonding "mm/internal/client/pumpfun/bonding"
	"mm/internal/client/raydium"
	"mm/internal/client/raydium/ammv4"
	"mm/internal/client/raydium/cpmm"
	"mm/internal/client/solanarpc"
	"mm/internal/client/solanaws"
	"mm/internal/cron"
	"mm/internal/crypto"
	"mm/internal/handler/server"
	v1 "mm/internal/handler/v1"
	"mm/internal/service"
	"mm/internal/storage/cache"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	"mm/internal/worker"
	auth "mm/pkg/jwt"
	"mm/pkg/logger"
	"mm/pkg/mailer"
	"mm/pkg/pool"
	"mm/pkg/validator"
	"time"

	"go.uber.org/fx"
)

func Build() *fx.App {
	return fx.New(
		fx.StartTimeout(900*time.Second),
		fx.StopTimeout(900*time.Second),

		fx.Options(
			config.Module,
			logger.Module,
			validator.Module,
			mailer.Module,
			pool.Module,
			jito.Module,
		),

		repository.Module(),
		secret.Module(),
		cache.Module(),

		service.Module(),

		server.Module(),

		v1.Module(),

		auth.Module(),

		cron.Module(),
		kucoinapi.Module(),
		solanarpc.Module(),
		solanaws.Module(),

		pumpfun.Module(),
		pumpAMM.Module(),
		pumpBonding.Module(),
		raydium.Module(),
		cpmm.Module(),
		ammv4.Module(),
		crypto.Module(),
		worker.Module(),
	)
}
