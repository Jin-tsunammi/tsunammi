package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("v1",
		fx.Provide(
			NewAccountHandler,
			NewDepositHandler,
			NewWalletHandler,
			NewProjectHandler,
			NewExchangeHandler,
			NewAuthHandler,
			NewUserHandler,
			NewSwapHandler,
			NewCampaignHandler,
			NewUtilHandler,
			NewSwaggerHandler,
			NewBuybackHandler,
			NewUploadHandler,
			NewPumpfunLaunchHandler,
		),
		fx.Invoke(func(app *fiber.App, accountHandler *AccountHandler, auth *AuthHandler) {
			accountHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, depositHandler *DepositHandler, auth *AuthHandler) {
			depositHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, walletHandler *WalletHandler, auth *AuthHandler) {
			walletHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, projectHandler *ProjectHandler, auth *AuthHandler) {
			projectHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, exchangeHandler *ExchangeHandler, auth *AuthHandler) {
			exchangeHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler) {
			authHandler.RegisterRoutes(app)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, userHandler *UserHandler) {
			userHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, swapHandler *SwapHandler, auth *AuthHandler) {
			swapHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, campaignHandler *CampaignHandler, auth *AuthHandler) {
			campaignHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, utilHandler *UtilHandler) {
			utilHandler.RegisterRoutes(app)
		}),
		fx.Invoke(func(app *fiber.App, buybackHandler *BuybackHandler, auth *AuthHandler) {
			buybackHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, swaggerHandler *SwaggerHandler) {
			swaggerHandler.RegisterRoutes(app)
		}),
		fx.Invoke(func(app *fiber.App, uploadHandler *UploadHandler, auth *AuthHandler) {
			uploadHandler.RegisterRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App, h *PumpfunLaunchHandler, auth *AuthHandler) {
			h.registerRoutes(app, auth)
		}),
		fx.Invoke(func(app *fiber.App) {
			app.Get("/static/*", static.New("./resources/static"))
		}),
	)
}
