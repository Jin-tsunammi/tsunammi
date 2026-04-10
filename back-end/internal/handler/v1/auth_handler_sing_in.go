package v1

import (
	"mm/internal/model"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"mm/pkg/mtype"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// SignInWithEmail godoc
//
//	@Summary		Sign in via email
//	@Description	Authenticates a user with an email and code, returning JWT access and refresh tokens.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.SignInWithEmail	true	"Sign In Request"
//	@Success		200		{object}	model.SignInResp		"Successful sign-in. Returns a user object and JWT info"
//	@Failure		400		{object}	apperrors.AppError		"Bad Request or invalid email format"
//	@Failure		401		{object}	apperrors.AppError		"Invalid confirmation code"
//	@Failure		404		{object}	apperrors.AppError		"User with this email not found"
//	@Failure		500		{object}	apperrors.AppError		"Internal Server Error"
//	@Router			/auth/sign-in-email [post]
func (h *AuthHandler) SignInWithEmail(c fiber.Ctx) error {
	var req model.SignInWithEmail
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	email, ok := mtype.NewEmail(req.Email)
	if !ok {
		return apperrors.BadRequest("invalid email")
	}

	ok, err := h.AuthService.CheckUsersEmailCode(c.Context(), email, req.Code)

	if err != nil {
		return apperrors.Unauthorized("failed to check users email code", err)
	}

	if !ok {
		return apperrors.Unauthorized("invalid code")
	}

	user, err := h.UserService.GetByEmail(c.Context(), email)
	if err != nil {
		return apperrors.NotFound("user with email not found", err)
	}

	claims := auth.NewTokenClaims(user.ID)
	tokenPair, err := h.JWTService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return err
	}

	jwtInfo := &model.SignInJWTResp{
		RefreshToken: tokenPair.RefreshToken,
		AccessToken:  tokenPair.AccessToken,
	}

	return c.Status(http.StatusOK).JSON(model.SignInResp{
		User:    user,
		JWTInfo: jwtInfo,
	})
}

// SignInWithGoogle godoc
//
//	@Summary		Sign in via Google
//	@Description	Authenticates an existing user via a Google ID token and returns JWT tokens.
//	@Tags			auth
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token (Google ID Token)"
//	@Success		200				{object}	model.SignInResp	"Successful sign-in. Returns a user object and JWT info"
//	@Failure		400				{object}	apperrors.AppError	"Email is missing from locals"
//	@Failure		401				{object}	apperrors.AppError	"Google authentication error"
//	@Failure		404				{object}	apperrors.AppError	"User not found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/auth/sign-in-google [post]
func (h *AuthHandler) SignInWithGoogle(c fiber.Ctx) error {
	email, ok := c.Locals("email").(mtype.Email)
	if !ok {
		return apperrors.BadRequest("email is missing")
	}

	user, err := h.UserService.GetByEmail(c.Context(), email)
	if err != nil {
		return apperrors.NotFound("user with email not found", err)
	}

	claims := auth.NewTokenClaims(user.ID)
	tokenPair, err := h.JWTService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return err
	}

	jwtInfo := &model.SignInJWTResp{
		RefreshToken: tokenPair.RefreshToken,
		AccessToken:  tokenPair.AccessToken,
	}

	return c.Status(http.StatusOK).JSON(model.SignInResp{
		User:    user,
		JWTInfo: jwtInfo,
	})
}

// SingInWithWallet godoc
//
//	@Summary		Sign in or register with Solana wallet
//	@Description	Authenticates a user by verifying a signed message from their Solana wallet. If the user does not exist, a new one is created.
//	@Tags			auth
//	@ID				sign-in-with-wallet
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.SolanaVerifyRequest	true	"Public key and signed message(public key) for verification"
//	@Success		200		{object}	model.SignInResp			"Successful sign-in or registration. Returns user data and JWT tokens"
//	@Failure		400		{object}	apperrors.AppError			"Bad Request (e.g., missing public key or signature)"
//	@Failure		401		{object}	apperrors.AppError			"Unauthorized (invalid wallet signature)"
//	@Failure		500		{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/auth/sign-in-wallet [post]
func (h *AuthHandler) SingInWithWallet(c fiber.Ctx) error {

	var req model.SolanaVerifyRequest

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("public key or singed message is missing")
	}

	_, err := h.UserService.VerifyAddress(c.Context(), req)
	if err != nil {
		return err
	}

	user, err := h.UserService.GetByPublicAddress(c.Context(), req.PublicAddress)

	if err != nil {
		return apperrors.Internal("cant authenticate user")
	}

	if user == nil {
		user, err = h.UserService.CreateWithPublicAddress(c.Context(), req.PublicAddress)
		if err != nil {
			return err
		}
	}

	claims := auth.NewTokenClaims(user.ID)
	tokenPair, err := h.JWTService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return err
	}

	jwtInfo := &model.SignInJWTResp{
		RefreshToken: tokenPair.RefreshToken,
		AccessToken:  tokenPair.AccessToken,
	}

	return c.Status(http.StatusOK).JSON(model.SignInResp{User: user, JWTInfo: jwtInfo})
}

// RefreshTokens godoc
//
//	@Summary		Refresh JWT tokens
//	@Description	Refreshes the access and refresh token pair using an existing refresh token.
//	@Tags			auth
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token (Refresh Token)"
//	@Success		200				{object}	auth.TokenPair		"Tokens successfully refreshed"
//	@Failure		401				{object}	apperrors.AppError	"Invalid or expired refresh token"
//	@Router			/auth/refresh [post]
//	@Security		BearerAuth
func (h *AuthHandler) RefreshTokens(c fiber.Ctx) error {
	token := c.Get("Authorization")

	tokenPair, err := h.JWTService.RefreshSession(c.Context(), token)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(tokenPair)
}
