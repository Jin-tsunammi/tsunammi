package auth

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("jwt",
		fx.Provide(
			fx.Annotate(
				NewJWTAuth,
				fx.As(
					new(JWTAuthenticator),
				),
			),
		),
	)
}
