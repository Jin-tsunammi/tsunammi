package crypto

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("crypto",
		fx.Provide(
			fx.Annotate(
				NewWalletEncryptor,
				fx.ResultTags(`name:"wallet_encryptor"`),
			),
			fx.Annotate(
				NewAccountEncryptor,
				fx.ResultTags(`name:"account_encryptor"`),
			),
		),
	)
}
