package solanaws

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("solana_ws",
		fx.Provide(
			NewClient,
		),
	)
}
