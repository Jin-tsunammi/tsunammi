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

type CampaignHandler struct {
	CampaignService *service.CampaignService
	Validator       *validator.Validate
}

func NewCampaignHandler(service *service.CampaignService, validate *validator.Validate) *CampaignHandler {
	return &CampaignHandler{
		CampaignService: service,
		Validator:       validate,
	}
}

// StopCampaign godoc
//
//	@Summary		Stop campaign
//	@Description	Stops an active campaign by its UUID.
//	@Tags			campaigns
//	@ID				stop-campaign
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Param			id				path		string				true	"Campaign UUID"	format(uuid)
//	@Success		204				{object}	nil					"No Content"
//	@Failure		400				{object}	apperrors.AppError	"Bad Request (invalid UUID)"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/campaigns/{id} [delete]
//	@Security		BearerAuth
func (h *CampaignHandler) StopCampaign(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	err = h.CampaignService.StopCampaign(c.Context(), uid, claims.UserID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}

// GetCampaigns godoc
//
//	@Summary		List campaigns
//	@Description	Returns all campaigns associated with the authenticated user.
//	@Tags			campaigns
//	@ID				get-campaigns
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Success		200				{array}		model.SwapCampaign
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/campaigns [get]
//	@Security		BearerAuth
func (h *CampaignHandler) GetCampaigns(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
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

	campaigns, err := h.CampaignService.GetCampaigns(c.Context(), claims.UserID, parsedPage, parsedPageSize)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(campaigns)
}

// UpdateCampaign godoc
//
//	@Summary		Update campaign
//	@Description	Updates configuration for a specific campaign.
//	@Tags			campaigns
//	@ID				update-campaign
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Authentication token"
//	@Param			id				path		string					true	"Campaign UUID"	format(uuid)
//	@Param			request			body		model.CampaignRequest	true	"Update parameters"
//	@Success		204				{object}	nil						"No Content"
//	@Failure		400				{object}	apperrors.AppError		"Bad Request"
//	@Failure		401				{object}	apperrors.AppError		"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError		"Internal Server Error"
//	@Router			/campaigns/{id} [put]
//	@Security		BearerAuth
func (h *CampaignHandler) UpdateCampaign(c fiber.Ctx) error {
	var req model.CampaignRequest

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	if err = c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("cant parse request", err)
	}

	if err = h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	err = h.CampaignService.UpdateCampaign(c.Context(), uid, claims.UserID, &req)

	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}

// GetCampaignByID godoc
//
//	@Summary		Get campaign by ID
//	@Description	Returns a specific campaign by its UUID for the authenticated user.
//	@Tags			campaigns
//	@ID				get-campaign-by-id
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			id				path		string	true	"Campaign UUID"	format(uuid)
//	@Success		200				{object}	model.SwapCampaign
//	@Failure		400				{object}	apperrors.AppError	"Bad Request (invalid UUID)"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/campaigns/{id} [get]
//	@Security		BearerAuth
func (h *CampaignHandler) GetCampaignByID(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	campaign, err := h.CampaignService.GetCampaignByID(c.Context(), uid, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(campaign)
}

// GetCampaignTransactions godoc
//
//	@Summary		Get campaign transactions
//	@Description	Returns a list of transactions associated with a specific campaign.
//	@Tags			campaigns
//	@ID				get-campaign-transactions
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			id				path		string	true	"Campaign UUID"	format(uuid)
//	@Success		200				{array}		model.SwapTransaction
//	@Failure		400				{object}	apperrors.AppError	"Bad Request (invalid UUID)"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/campaigns/{id}/transactions [get]
//	@Security		BearerAuth
func (h *CampaignHandler) GetCampaignTransactions(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
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

	transactions, err := h.CampaignService.GetCampaignTransactions(c.Context(), uid, claims.UserID, parsedPage, parsedPageSize)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(transactions)
}

// GetCampaignsSummary godoc
//
//	@Summary		Get campaigns summary
//	@Description	Returns a statistical summary of all campaigns for the authenticated user (e.g., total spent, active count).
//	@Tags			campaigns
//	@ID				get-campaigns-summary
//	@Produce		json
//	@Param			Authorization	header		string					true	"Authentication token"
//	@Success		200				{array}		model.CampaignSummary	"Campaigns summary data"
//	@Failure		401				{object}	apperrors.AppError		"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError		"Internal Server Error"
//	@Router			/campaigns/summary [get]
//	@Security		BearerAuth
func (h *CampaignHandler) GetCampaignsSummary(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
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

	status := c.Query("status")
	campaignType := c.Query("type")

	var parsedStatus, parsedCampaignType string

	switch status {
	case model.StatusStop:
		parsedStatus = model.StatusStop
	case model.StatusInUse:
		parsedStatus = model.StatusInUse
	case model.StatusDone:
		parsedStatus = model.StatusDone
	case "":
		parsedStatus = ""
	default:
		return apperrors.BadRequest("status must be either '', 'stop', 'in_use' or 'done'")
	}

	switch campaignType {
	case "pull_up":
		parsedCampaignType = model.TargetUpTaskType
	case "pull_down":
		parsedCampaignType = model.TargetDownTaskType
	case "":
		parsedCampaignType = ""
	default:
		return apperrors.BadRequest("type must be either '', 'pull_up' or 'pull_down'")

	}

	summary, err := h.CampaignService.GetCampaignsSummary(c.Context(), parsedPage, parsedPageSize, claims.UserID, parsedCampaignType, parsedStatus)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(summary)
}

func (h *CampaignHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	campaignGroup := app.Group("/campaigns")
	campaignGroup.Use(auth.AuthMiddleware)
	{
		campaignGroup.Get("", h.GetCampaigns)
		campaignGroup.Get("/:id", h.GetCampaignByID)
		campaignGroup.Patch("/:id", h.UpdateCampaign)
		campaignGroup.Delete("/:id", h.StopCampaign)
		campaignGroup.Get("/:id/transactions", h.GetCampaignTransactions)
	}

	campaignSummaryGroup := app.Group("/campaigns-summary")
	campaignSummaryGroup.Use(auth.AuthMiddleware)
	{
		campaignSummaryGroup.Get("", h.GetCampaignsSummary)
	}
}
