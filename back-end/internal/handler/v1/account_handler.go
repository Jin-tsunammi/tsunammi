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

type AccountHandler struct {
	AccountService *service.AccountService
	Validator      *validator.Validate
}

func NewAccountHandler(service *service.AccountService, validate *validator.Validate) *AccountHandler {
	return &AccountHandler{
		AccountService: service,
		Validator:      validate,
	}
}

// AddExchangeAccount
//
//	@Summary		Create exchange account
//	@Description	Creates a new exchange account
//	@Tags			accounts
//	@ID				add-exchange-account
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			account			body		model.AddExchangeAccountReq	true	"Payload to create an account"
//	@Success		201				{object}	model.Account				"Created"
//	@Failure		400				{object}	apperrors.AppError			"Bad Request"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/accounts [post]
//	@Security		BearerAuth
func (h *AccountHandler) AddExchangeAccount(c fiber.Ctx) error {
	var req model.AddExchangeAccountReq

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

	res, err := h.AccountService.AddExchangeAccount(c.Context(), &req, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(res)
}

// GetAccountsByUserID godoc
//
//	@Summary		Get all accounts
//	@Description	Returns all exchange accounts for the authenticated user.
//	@Tags			accounts
//	@ID				get-accounts-by-user
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Success		200				{array}		model.AccountResponse
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/accounts [get]
//	@Security		BearerAuth
func (h *AccountHandler) GetAccountsByUserID(c fiber.Ctx) error {

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

	accounts, err := h.AccountService.GetAccountsByUserID(c.Context(), parsedPage, parsedPageSize, claims.UserID)

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(accounts)
}

// GetAccountByIDAndUserID godoc
//
//	@Summary		Get account by ID
//	@Description	Returns a specific exchange account by its ID, for the authenticated user.
//	@Tags			accounts
//	@ID				get-account-by-id
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			id				path		int		true	"Account ID"
//	@Success		200				{object}	model.AccountResponse
//	@Failure		400				{object}	apperrors.AppError	"Bad Request (invalid ID)"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/accounts/{id} [get]
//	@Security		BearerAuth
func (h *AccountHandler) GetAccountByIDAndUserID(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	account, err := h.AccountService.GetAccountByIDAndUserID(c.Context(), id, claims.UserID)

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(account)
}

// DeleteAccountByIDAndUserID godoc
//
//	@Summary		Delete account by ID
//	@Description	Deletes a specific exchange account by its ID, for the authenticated user.
//	@Tags			accounts
//	@ID				delete-account-by-id
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Param			id				path		int					true	"Account ID"
//	@Success		204				{object}	nil					"No Content"
//	@Failure		400				{object}	apperrors.AppError	"Bad Request (invalid ID)"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/accounts/{id} [delete]
//	@Security		BearerAuth
func (h *AccountHandler) DeleteAccountByIDAndUserID(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	err = h.AccountService.DeleteAccountByIDAndUserID(c.Context(), id, claims.UserID)

	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}

func (h *AccountHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	accounts := app.Group("/accounts")
	accounts.Use(auth.AuthMiddleware)
	{
		accounts.Post("", h.AddExchangeAccount)
		accounts.Get("", h.GetAccountsByUserID)
		accounts.Get("/:id", h.GetAccountByIDAndUserID)
		accounts.Delete("/:id", h.DeleteAccountByIDAndUserID)
	}
}
