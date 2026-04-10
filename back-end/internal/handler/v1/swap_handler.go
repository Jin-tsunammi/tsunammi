package v1

import (
	"mm/internal/client/jito"
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type SwapHandler struct {
	Validator     *validator.Validate
	PullUpService *service.SwapService
	JitoClient    *jito.Client
}

func NewSwapHandler(
	pullUpService *service.SwapService,
	jitoClient *jito.Client,
	validate *validator.Validate,
) *SwapHandler {
	return &SwapHandler{
		PullUpService: pullUpService,
		JitoClient:    jitoClient,
		Validator:     validate,
	}
}

// GetTipFloor godoc
//
//	@Summary		Get Jito tip floor
//	@Description	Returns the current tip floor information from Jito.
//	@Tags			jito
//	@ID				get-jito-tip-floor
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Success		200				{object}	model.TipFloorSOL	"Tip floor data"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/jito/tip-floor [get]
//	@Security		BearerAuth
func (h *SwapHandler) GetTipFloor(c fiber.Ctx) error {
	_, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	tipFloor, err := h.JitoClient.GetTipFloor(c.Context())

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(
		model.TipFloorSOL{
			Default: tipFloor.LandedTips75ThPercentile,
			Fast:    tipFloor.LandedTips95ThPercentile,
			Extra:   tipFloor.LandedTips99ThPercentile,
		})

}

// SwapTargetPullUp godoc
//
//	@Summary		Create Pull Up campaign (Raydium)
//	@Description	Creates a new target pull-up campaign on Raydium to increase token price.
//	@Tags			raydium
//	@ID				raydium-target-pull-up
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			request			body		model.TargetPullUpRequest	true	"Pull Up parameters"
//	@Success		201				{object}	model.TargetPullResponse	"Created"
//	@Failure		400				{object}	apperrors.AppError			"Bad Request"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/raydium/pull-up [post]
//	@Security		BearerAuth
func (h *SwapHandler) SwapTargetPullUp(c fiber.Ctx) error {
	return h.swapTargetPullUp(c, model.SwapProviderRaydium)
}

func (h *SwapHandler) swapTargetPullUp(c fiber.Ctx, providerID model.SwapProviderID) error {

	var req model.TargetPullUpRequest

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("cant parse request "+err.Error(), err)
	}

	req.ProviderID = providerID

	if err := h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	res, err := h.PullUpService.CreatePullUpCampaign(c.Context(), &req, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(res)

}

// PumpfunTargetPullUp godoc
//
//	@Summary		Create Pull Up campaign (Pumpfun)
//	@Description	Creates a new target pull-up campaign on Pumpfun to increase token price.
//	@Tags			pumpfun
//	@ID				pumpfun-target-pull-up
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			request			body		model.TargetPullUpRequest	true	"Pull Up parameters"
//	@Success		201				{object}	model.TargetPullResponse	"Created"
//	@Failure		400				{object}	apperrors.AppError			"Bad Request"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/pumpfun/pull-up [post]
//	@Security		BearerAuth
func (h *SwapHandler) PumpfunTargetPullUp(c fiber.Ctx) error {
	return h.swapTargetPullUp(c, model.SwapProviderPumpfun)
}

// SwapTargetPullDown godoc
//
//	@Summary		Create Pull Down campaign (Raydium)
//	@Description	Creates a new target pull-down campaign on Raydium to decrease token price.
//	@Tags			raydium
//	@ID				raydium-target-pull-down
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			request			body		model.TargetPullDownRequest	true	"Pull Down parameters"
//	@Success		201				{object}	model.TargetPullResponse	"Created"
//	@Failure		400				{object}	apperrors.AppError			"Bad Request"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/raydium/pull-down [post]
//	@Security		BearerAuth
func (h *SwapHandler) SwapTargetPullDown(c fiber.Ctx) error {
	return h.swapTargetPullDown(c, model.SwapProviderRaydium)
}

func (h *SwapHandler) swapTargetPullDown(c fiber.Ctx, providerID model.SwapProviderID) error {

	var req model.TargetPullDownRequest

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("cant parse request", err)
	}

	req.ProviderID = providerID

	if err := h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	res, err := h.PullUpService.CreatePullDownCampaign(c.Context(), &req, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(res)

}

// PumpfunTargetPullDown godoc
//
//	@Summary		Create Pull Down campaign (Pumpfun)
//	@Description	Creates a new target pull-down campaign on Pumpfun to decrease token price.
//	@Tags			pumpfun
//	@ID				pumpfun-target-pull-down
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			request			body		model.TargetPullDownRequest	true	"Pull Down parameters"
//	@Success		201				{object}	model.TargetPullResponse	"Created"
//	@Failure		400				{object}	apperrors.AppError			"Bad Request"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/pumpfun/pull-down [post]
//	@Security		BearerAuth
func (h *SwapHandler) PumpfunTargetPullDown(c fiber.Ctx) error {
	return h.swapTargetPullDown(c, model.SwapProviderPumpfun)
}

// SwapEstimatePull godoc
//
//	@Summary		Estimate Raydium Pull
//	@Description	Estimates the cost and impact of a Raydium pull operation.
//	@Tags			raydium
//	@ID				raydium-estimate-pull
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string								true	"Authentication token"
//	@Param			request			body		model.EstimatePullRequest			true	"Estimate parameters"
//	@Success		200				{object}	model.TargetPullEstimateResponse	"Estimation Result"
//	@Failure		400				{object}	apperrors.AppError					"Bad Request"
//	@Failure		401				{object}	apperrors.AppError					"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError					"Internal Server Error"
//	@Router			/raydium/estimate [post]
//	@Security		BearerAuth
func (h *SwapHandler) SwapEstimatePull(c fiber.Ctx) error {
	return h.swapEstimatePull(c, model.SwapProviderRaydium)
}

func (h *SwapHandler) swapEstimatePull(c fiber.Ctx, providerID model.SwapProviderID) error {

	var req model.EstimatePullRequest

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest(err.Error(), err)
	}

	req.ProviderID = providerID

	if err := h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	res, err := h.PullUpService.EstimateSwapCost(c.Context(), &req, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)

}

// PumpfunEstimatePull godoc
//
//	@Summary		Estimate Pumpfun Pull
//	@Description	Estimates the cost and impact of a Pumpfun pull operation.
//	@Tags			pumpfun
//	@ID				pumpfun-estimate-pull
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string								true	"Authentication token"
//	@Param			request			body		model.EstimatePullRequest			true	"Estimate parameters"
//	@Success		200				{object}	model.TargetPullEstimateResponse	"Estimation Result"
//	@Failure		400				{object}	apperrors.AppError					"Bad Request"
//	@Failure		401				{object}	apperrors.AppError					"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError					"Internal Server Error"
//	@Router			/pumpfun/estimate [post]
//	@Security		BearerAuth
func (h *SwapHandler) PumpfunEstimatePull(c fiber.Ctx) error {
	return h.swapEstimatePull(c, model.SwapProviderPumpfun)
}

func (h *SwapHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	raydiumGroup := app.Group("/raydium")
	raydiumGroup.Use(auth.AuthMiddleware)
	{
		raydiumGroup.Post("/pull-up", h.SwapTargetPullUp)
		raydiumGroup.Post("/pull-down", h.SwapTargetPullDown)
		raydiumGroup.Post("/estimate", h.SwapEstimatePull)
	}

	pumpfunGroup := app.Group("/pumpfun")
	pumpfunGroup.Use(auth.AuthMiddleware)
	{
		pumpfunGroup.Post("/pull-up", h.PumpfunTargetPullUp)
		pumpfunGroup.Post("/pull-down", h.PumpfunTargetPullDown)
		pumpfunGroup.Post("/estimate", h.PumpfunEstimatePull)
	}

	jitoGroup := app.Group("/jito")
	jitoGroup.Use(auth.AuthMiddleware)
	jitoGroup.Get("/tip-floor", h.GetTipFloor)
}
