package service

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("service",
		fx.Provide(
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
		),
	)
}
