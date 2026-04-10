package server

import (
	"context"
	"log"
	"mm/config"
	"net"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("handler",
		fx.Provide(
			NewServer,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, app *fiber.App, c *config.Config) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							addr := net.JoinHostPort(c.HTTP.Host, c.HTTP.Port)
							if err := app.Listen(addr); err != nil {
								log.Println("failed to start handler listening", err)
							}
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						return app.ShutdownWithContext(ctx)
					},
				},
				)
			},
		),
	)
}
