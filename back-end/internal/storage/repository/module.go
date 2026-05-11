package repository

import (
	"context"
	"mm/internal/model"
	"mm/pkg/repository"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("db",
		fx.Provide(
			CreateDBConnection,
		),
		fx.Provide(
			repository.NewTransactionManager,
		),
		fx.Provide(
			fx.Annotate(
				repository.NewDBWrapper,
				fx.As(
					new(repository.DB),
				)),
		),
		fx.Provide(
			repository.NewGenericRepository[model.Account, uint64],
			NewAccountRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Exchange, uint64],
			NewExchangeRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Project, uint64],
			NewProjectRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Wallet, uint64],
			NewWalletRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Deposit, uint64],
			NewDepositRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.DepositOrder, uint64],
			NewDepositOrderRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.EmailVerificationCode, uint64],
			fx.Annotate(
				NewCodeRepository,
				fx.As(
					new(CodeRepository),
				)),
		),
		fx.Provide(
			repository.NewGenericRepository[model.Session, uuid.UUID],
			fx.Annotate(
				NewJWTRepository,
				fx.As(
					new(JWTRepository),
				)),
		),
		fx.Provide(
			repository.NewGenericRepository[model.UserHistory, uint64],

			NewUserHistoryRepository,
		),

		fx.Provide(
			repository.NewGenericRepository[model.User, uint64],
			NewUserRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.SwapTransaction, uint64],
			NewSwapTransactionRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.SwapCampaign, uint64],
			NewSwapCampaignRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.SmartBuybackCampaign, uuid.UUID],
			NewBuybackRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.BuybackTransaction, uint64],
			NewBuybackTransactionRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.PumpfunLaunch, uuid.UUID],
			NewPumpfunLaunchRepository,
		),

		fx.Invoke(
			func(lc fx.Lifecycle, db *bun.DB) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						return db.Close()
					},
				},
				)
			},
		),
	)
}
