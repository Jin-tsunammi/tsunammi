package v1

import (
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type DepositHandler struct {
	Validator      *validator.Validate
	DepositService *service.DepositService
}

func NewDepositHandler(service *service.DepositService, validate *validator.Validate) *DepositHandler {
	return &DepositHandler{
		DepositService: service,
		Validator:      validate,
	}
}

// DepositSolana godoc
//
//	@Summary		Create Solana deposit
//	@Description	Creates a new Solana deposit request
//	@Tags			deposits
//	@ID				deposit-solana
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Authentication token"
//	@Param			deposit			body		model.DepositSolanaReq	true	"Deposit request payload"
//	@Success		201				{object}	model.DepositResponse	"Created"
//	@Failure		400				{object}	apperrors.AppError		"Bad Request"
//	@Failure		401				{object}	apperrors.AppError		"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError		"Not Found"
//	@Failure		500				{object}	apperrors.AppError		"Internal Server Error"
//	@Router			/wallets/solana/deposit [post]
//	@Security		BearerAuth
func (h *DepositHandler) DepositSolana(c fiber.Ctx) error {
	var req model.DepositSolanaReq

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("cant parse request", err)
	}

	if err := h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	res, err := h.DepositService.DepositSolana(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(res)
}

// ProcessDeposit godoc
//
//	@Summary		Process a deposit order
//	@Description	Processes a previously created deposit order by its ID.
//	@Tags			deposits
//	@ID				process-deposit-order
//	@Produce		json
//	@Param			Authorization	header		string							true	"Authentication token"
//	@Param			id				path		int								true	"Deposit Order ID"
//	@Success		202				{object}	model.DepositProcessResponse	"Accepted"
//	@Failure		400				{object}	apperrors.AppError				"Bad Request (invalid ID)"
//	@Failure		401				{object}	apperrors.AppError				"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError				"Not Found"
//	@Failure		500				{object}	apperrors.AppError				"Internal Server Error"
//	@Router			/wallets/solana/deposit/process/{id} [post]
//	@Security		BearerAuth
func (h *DepositHandler) ProcessDeposit(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	order, err := h.DepositService.ProcessDepositOrder(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}

	return c.Status(http.StatusAccepted).JSON(order)

}

// GetDepositStatus godoc
//
//	@Summary		Get deposit status
//	@Description	Returns the current status of a deposit by its ID
//	@Tags			deposits
//	@ID				get-deposit-status
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			id				path		int		true	"Deposit ID"
//	@Success		200				{object}	model.Deposit
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/wallets/solana/deposit/{id} [post]
//	@Security		BearerAuth
func (h *DepositHandler) GetDepositStatus(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	res, err := h.DepositService.GetDepositStatus(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

// GetDepositHistory godoc
//
//	@Summary		Get deposit history
//	@Description	Returns the deposit history for the authenticated user.
//	@Tags			deposits
//	@ID				get-deposit-history
//	@Produce		json
//	@Param			Authorization	header		string							true	"Authentication token"
//	@Success		200				{array}		model.DepositHistoryResponse	"OK"
//	@Failure		401				{object}	apperrors.AppError				"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError				"Internal Server Error"
//	@Router			/wallets/solana/deposit/history [get]
//	@Security		BearerAuth
func (h *DepositHandler) GetDepositHistory(c fiber.Ctx) error {

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

	history, err := h.DepositService.GetDepositHistory(c.Context(), claims.UserID, parsedPage, parsedPageSize)

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(history)
}

// GetDepositHistoryByProjectID godoc
//
//	@Summary		Get deposit history by project ID
//	@Description	Returns the deposit history for a specific project ID, owned by the authenticated user.
//	@Tags			deposits
//	@ID				get-deposit-history-by-project-id
//	@Produce		json
//	@Param			Authorization	header		string							true	"Authentication token"
//	@Param			id				path		int								true	"Project ID"
//	@Success		200				{object}	model.DepositHistoryResponse	"OK"
//	@Failure		400				{object}	apperrors.AppError				"Bad Request (invalid ID)"
//	@Failure		401				{object}	apperrors.AppError				"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError				"Not Found"
//	@Failure		500				{object}	apperrors.AppError				"Internal Server Error"
//	@Router			/wallets/solana/deposit/history/{id} [get]
//	@Security		BearerAuth
func (h *DepositHandler) GetDepositHistoryByProjectID(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	history, err := h.DepositService.GetDepositHistoryByProjectID(c.Context(), id, claims.UserID)

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(history)
}

func (h *DepositHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	wallets := app.Group("/wallets")
	{
		solana := wallets.Group("/solana")
		solana.Use(auth.AuthMiddleware)
		{
			solana.Post("/deposit", h.DepositSolana)
			solana.Post("/deposit/process/:id", h.ProcessDeposit)
			solana.Post("/deposit/:id", h.GetDepositStatus)
			solana.Get("/deposit/history", h.GetDepositHistory)
			solana.Get("/deposit/history/:id", h.GetDepositHistoryByProjectID)
		}
	}
}
