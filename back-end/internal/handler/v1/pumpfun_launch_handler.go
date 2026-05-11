package v1

import (
	"encoding/json"
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

type PumpfunLaunchHandler struct {
	pumpfunService *service.PumpfunLaunchService
}

func NewPumpfunLaunchHandler(service *service.PumpfunLaunchService) *PumpfunLaunchHandler {
	return &PumpfunLaunchHandler{
		pumpfunService: service,
	}
}

// prepareCreateTx godoc
//
//	@Summary		Prepare Pumpfun token launch
//	@Description	Uploads token metadata, builds the unsigned Pumpfun create transaction and optional initial buy transactions, then stores pending launch state.
//	@Tags			pumpfun-launch
//	@ID				prepare-pumpfun-launch
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			logo			formData	file	true	"Token logo image"
//	@Param			data			formData	string	true	"JSON encoded model.PumpfunPrepareCreateTxRequest without logo"
//	@Success		200				{object}	model.PumpfunPrepareCreateTxResponse
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		422				{object}	apperrors.AppError	"Unprocessable Entity"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/launch/pumpfun/prepare [post]
//	@Security		BearerAuth
func (h *PumpfunLaunchHandler) prepareCreateTx(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	file, err := c.FormFile("logo")
	if err != nil {
		return apperrors.BadRequest("logo is required")
	}

	var req model.PumpfunPrepareCreateTxRequest
	if err := json.Unmarshal([]byte(c.FormValue("data")), &req); err != nil {
		return apperrors.BadRequest("invalid request")
	}

	req.Logo = file

	res, err := h.pumpfunService.PrepareCreateTx(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

// estimateCreateTx godoc
//
//	@Summary		Estimate Pumpfun token launch costs
//	@Description	Returns total expected token amount, hardcoded creation fee in SOL, configured Jito tip in SOL, total priority fee in SOL, and all remaining Pumpfun commissions as one SOL value.
//	@Tags			pumpfun-launch
//	@ID				estimate-pumpfun-launch
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string								true	"Authentication token"
//	@Param			estimate		body		model.PumpfunEstimateCreateRequest	true	"Pumpfun launch estimate payload"
//	@Success		200				{object}	model.PumpfunEstimateCreateResponse
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/launch/pumpfun/estimate [post]
//	@Security		BearerAuth
func (h *PumpfunLaunchHandler) estimateCreateTx(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.PumpfunEstimateCreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.BadRequest("invalid request", err)
	}

	res, err := h.pumpfunService.EstimateCreateTx(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

// processCreate godoc
//
//	@Summary		Launch prepared Pumpfun token
//	@Description	Verifies signed launch transactions for a pending Pumpfun launch and broadcasts them as a bundle.
//	@Tags			pumpfun-launch
//	@ID				process-pumpfun-launch
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string								true	"Authentication token"
//	@Param			launch			body		model.PumpfunProcessCreateRequest	true	"Signed launch transaction payload"
//	@Success		200				{object}	model.PumpfunProcessCreateResponse
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		422				{object}	apperrors.AppError	"Unprocessable Entity"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/launch/pumpfun/launch [post]
//	@Security		BearerAuth
func (h *PumpfunLaunchHandler) processCreate(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.PumpfunProcessCreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.BadRequest("invalid request", err)
	}

	res, err := h.pumpfunService.Launch(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *PumpfunLaunchHandler) registerRoutes(app *fiber.App, auth *AuthHandler) {
	group := app.Group("/launch/pumpfun")
	group.Use(auth.AuthMiddleware)
	{
		group.Post("/estimate", h.estimateCreateTx)
		group.Post("/prepare", h.prepareCreateTx)
		group.Post("/launch", h.processCreate)
	}
}
