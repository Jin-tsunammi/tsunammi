package v1

import (
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type ExchangeHandler struct {
	Validator       *validator.Validate
	ExchangeService *service.ExchangeService
}

func NewExchangeHandler(service *service.ExchangeService, validate *validator.Validate) *ExchangeHandler {
	return &ExchangeHandler{
		ExchangeService: service,
		Validator:       validate,
	}
}

// GetExchanges godoc
//
//	@Summary		List exchanges
//	@Description	Returns all exchanges
//	@Tags			exchanges
//	@ID				get-exchanges
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Success		200				{array}		model.Exchange
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/exchanges [get]
//	@Security		BearerAuth
func (h *ExchangeHandler) GetExchanges(c fiber.Ctx) error {

	_, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	res, err := h.ExchangeService.GetExchanges(c.Context())
	if err != nil {
		return apperrors.Internal("cant get exchanges", err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *ExchangeHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	exchanges := app.Group("/exchanges")
	exchanges.Use(auth.AuthMiddleware)
	{
		exchanges.Get("", h.GetExchanges)
	}
}
