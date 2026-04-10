package v1

import (
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"mm/pkg/mtype"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	UserService *service.UserService
	AuthService *service.AuthService
}

func NewUserHandler(
	userService *service.UserService,
	authService *service.AuthService,
) *UserHandler {
	return &UserHandler{
		UserService: userService,
		AuthService: authService,
	}
}

// GetUser godoc
//
//	@Summary		Get current user info
//	@Description	Returns the profile data for the user authenticated via a JWT access token.
//	@Tags			user
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Success		200				{object}	model.User			"User data"
//	@Failure		401				{object}	apperrors.AppError	"Authentication error, claims not found"
//	@Failure		404				{object}	apperrors.AppError	"User not found"
//	@Router			/user [get]
//	@Security		BearerAuth
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	user, err := h.UserService.GetByID(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(user)
}

// GetUserHistory godoc
//
//	@Summary		Get user action history
//	@Description	Returns a list of actions (history) for the authenticated user.
//	@Tags			user
//	@ID				get-user-history
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Success		200				{array}		model.UserHistory	"User action history"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/user/history [get]
//	@Security		BearerAuth
func (h *UserHandler) GetUserHistory(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	from := c.Query("from")
	to := c.Query("to")

	fromParsed := time.Time{}
	toParsed := time.Time{}

	var err error

	if from != "" && to != "" {
		fromParsed, err = time.Parse(time.RFC3339, from)
		if err != nil {
			return apperrors.BadRequest("invalid from date format")
		}

		toParsed, err = time.Parse(time.RFC3339, to)
		if err != nil {
			return apperrors.BadRequest("invalid to date format")
		}
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

	actions, err := h.UserService.GetHistoryByUserID(c.Context(), claims.UserID, parsedPage, parsedPageSize, fromParsed, toParsed)

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(actions)
}

// UpdateUserEmail godoc
//
//	@Summary		Update user email
//	@Description	Updates the email address of the authenticated user. Requires a valid confirmation code sent to the new email.
//	@Tags			user
//	@ID				update-user-email
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			request			body		model.ChangeEmailRequest	true	"New email and verification code"
//	@Success		200				{string}	string						"OK"
//	@Failure		400				{object}	apperrors.AppError			"Invalid request data or email format"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized or invalid verification code"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/user/email [patch]
//	@Security		BearerAuth
func (h *UserHandler) UpdateUserEmail(c fiber.Ctx) error {
	var req model.ChangeEmailRequest

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	email, ok := mtype.NewEmail(req.Email)
	if !ok {
		return apperrors.BadRequest("invalid email")
	}

	ok, err := h.AuthService.CheckUsersEmailCode(c.Context(), email, req.Code)
	if err != nil {
		return err
	}

	if !ok {
		return apperrors.Unauthorized("invalid code")
	}

	err = h.UserService.ChangeUserEmail(c.Context(), email, claims.UserID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (h *UserHandler) DebugDeleteUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	err := h.UserService.DebugDeleteUser(c.Context(), claims.UserID)

	if err != nil {
		return apperrors.Internal("cant delete user", err)
	}

	return c.SendStatus(http.StatusNoContent)
}

func (h *UserHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	userGroup := app.Group("/user")

	userGroup.Use(auth.AuthMiddleware)
	{
		userGroup.Get("", h.GetUser)
		userGroup.Get("/history", h.GetUserHistory)
		userGroup.Patch("/email", h.UpdateUserEmail)
		userGroup.Delete("/debug", h.DebugDeleteUser)
	}

}
