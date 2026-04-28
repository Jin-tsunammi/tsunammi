package v1

import (
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type BuybackHandler struct {
	BuybackService *service.SmartBuybackService
	Validator      *validator.Validate
}

func NewBuybackHandler(service *service.SmartBuybackService, validator *validator.Validate) *BuybackHandler {
	return &BuybackHandler{
		BuybackService: service,
		Validator:      validator,
	}
}

func (h *BuybackHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	buybackGroup := app.Group("/buyback")
	buybackGroup.Use(auth.AuthMiddleware)
	{
		buybackGroup.Get("", h.getAll)
		buybackGroup.Get("/:id", h.getById)
		buybackGroup.Post("", h.create)
		buybackGroup.Delete("/:id", h.stop)
		buybackGroup.Get("/:id/transactions", h.getTransactions)
	}
}

// create godoc
//
//	@Summary		Create buyback campaign
//	@Description	Creates a new smart buyback campaign with one or more targets.
//	@Tags			buyback
//	@ID				create-buyback-campaign
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string									true	"Authentication token"
//	@Param			request			body		model.CreateSmartBuybackCampaignRequest	true	"Campaign parameters"
//	@Success		200				{object}	model.SmartBuybackCampaignWithTargets
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/buyback [post]
//	@Security		BearerAuth
func (h *BuybackHandler) create(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.CreateSmartBuybackCampaignRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.BadRequest("invalid request", err)
	}

	if err := h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	res, err := h.BuybackService.CreateCampaign(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

// getAll godoc
//
//	@Summary		List buyback campaigns
//	@Description	Returns all buyback campaigns for the authenticated user.
//	@Tags			buyback
//	@ID				get-buyback-campaigns
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Success		200				{array}		model.SmartBuybackCampaign
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/buyback [get]
//	@Security		BearerAuth
func (h *BuybackHandler) getAll(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	res, err := h.BuybackService.GetCampaigns(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

// getById godoc
//
//	@Summary		Get buyback campaign by ID
//	@Description	Returns a buyback campaign with its targets by UUID.
//	@Tags			buyback
//	@ID				get-buyback-campaign-by-id
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			id				path		string	true	"Campaign UUID"	format(uuid)
//	@Success		200				{object}	model.SmartBuybackCampaignWithTargets
//	@Failure		400				{object}	apperrors.AppError	"Bad Request (invalid UUID)"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/buyback/{id} [get]
//	@Security		BearerAuth
func (h *BuybackHandler) getById(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid id", err)
	}

	res, err := h.BuybackService.GetByID(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

// stop godoc
//
//	@Summary		Stop buyback campaign
//	@Description	Stops an active buyback campaign by its UUID.
//	@Tags			buyback
//	@ID				stop-buyback-campaign
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Param			id				path		string				true	"Campaign UUID"	format(uuid)
//	@Success		204				{object}	nil					"No Content"
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/buyback/{id} [delete]
//	@Security		BearerAuth
func (h *BuybackHandler) stop(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid id", err)
	}

	if err := h.BuybackService.StopCampaign(c.Context(), claims.UserID, id); err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}

// getTransactions godoc
//
//	@Summary		Get buyback transactions
//	@Description	Returns a paginated list of transactions for a buyback campaign. Optionally filtered by target ID.
//	@Tags			buyback
//	@ID				get-buyback-transactions
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			id				path		string	false	"Campaign UUID"			format(uuid)
//	@Param			target_id		query		string	false	"Target UUID filter"	format(uuid)
//	@Param			page			query		int		false	"Page number"
//	@Param			pageSize		query		int		false	"Page size"
//	@Success		200				{array}		model.BuybackTransaction
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/buyback/{id}/transactions [get]
//	@Security		BearerAuth
func (h *BuybackHandler) getTransactions(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid id", err)
	}

	page := c.Query("page")
	pageSize := c.Query("pageSize")

	if (page == "" && pageSize != "") || (page != "" && pageSize == "") {
		return apperrors.BadRequest("query params 'page' and 'pageSize' must be provided together. One is missing.")
	}

	parsedPage, parsedPageSize, err := parsePaginationParams(page, pageSize)
	if err != nil {
		return err
	}

	targetID := uuid.Nil
	if t := c.Query("target_id"); t != "" {
		targetID, _ = uuid.Parse(t)
	}

	res, err := h.BuybackService.GetTransactions(c.Context(), id, claims.UserID, targetID, parsedPage, parsedPageSize)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}
