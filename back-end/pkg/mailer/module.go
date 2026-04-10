package mailer

import (
	"go.uber.org/fx"
)

var Module = fx.Module("mailer",
	fx.Provide(
		fx.Annotate(
			NewMailer,
			fx.As(
				new(Mailer),
			)),
	),
)
