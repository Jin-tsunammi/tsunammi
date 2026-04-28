package service

import (
	"mm/internal/dex"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("service",
		fx.Provide(
			dex.NewDexProviders,
			NewExchangeService,
			NewProjectService,
			NewWalletService,
			NewDepositService,
			NewAccountService,
			NewAuthService,
			NewUserService,
			NewJWTService,
			NewSwapService,
			NewCampaignService,
			NewSmartBuybackService,
		),
	)
}
