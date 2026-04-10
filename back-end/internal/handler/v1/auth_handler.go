package v1

import (
	"context"
	"fmt"
	"mm/config"
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"mm/pkg/mtype"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/api/option"
)

const (
	jwtAuthType = "Bearer"
)

type AuthHandler struct {
	Firebase    *firebaseAuth.Client
	JWTService  *service.JWTService
	AuthService *service.AuthService
	UserService *service.UserService
}

func NewAuthHandler(
	c *config.Config,
	userService *service.UserService,
	authService *service.AuthService,
	jwtService *service.JWTService,
) (*AuthHandler, error) {
	authHandler := &AuthHandler{
		UserService: userService,
		JWTService:  jwtService,
		AuthService: authService,
	}

	err := authHandler.setFirebaseAuth(c.App.FirebaseFilePath)
	if err != nil {
		return nil, err
	}

	return authHandler, nil
}

func (h *AuthHandler) setFirebaseAuth(firebaseFilePath string) error {
	opt := option.WithCredentialsFile(firebaseFilePath)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("firebase.NewApp: %w", err)
	}

	h.Firebase, err = app.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("app.Auth: %w", err)
	}

	return nil
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	authGroup := app.Group("/auth")
	{
		authGroup.Post("/send-code", h.SendCode)
		authGroup.Post("/is-user-exists", h.IsUserExists)

		authGroup.Post("/sign-up-email", h.SignUpWithEmail)
		authGroup.Post("/sign-up-google", h.SignUpWithGoogleMiddleware, h.SignUpWithGoogle)

		authGroup.Post("/sign-in-email", h.SignInWithEmail)
		authGroup.Post("/sign-in-google", h.SignInWithGoogleMiddleware, h.SignInWithGoogle)

		authGroup.Post("/sign-in-wallet", h.SingInWithWallet)

		authGroup.Post("/refresh", h.RefreshTokens)
	}
}

func (h *AuthHandler) IsUserExists(c fiber.Ctx) error {

	var req model.IsUserExistReq

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	email, ok := mtype.NewEmail(req.Email)

	if !ok {
		return apperrors.BadRequest("invalid email")
	}

	_, err := h.UserService.GetByEmail(c.Context(), email)

	resp := model.IsUserExistResp{
		Exist: true,
	}

	if err != nil {
		resp = model.IsUserExistResp{
			Exist: false,
		}
	}

	return c.Status(http.StatusOK).JSON(resp)
}

func (h *AuthHandler) SignUpWithGoogleMiddleware(c fiber.Ctx) error {
	if err := h.authWithGoogle(c); err != nil {
		return err
	}

	return c.Next()
}

func (h *AuthHandler) SignInWithGoogleMiddleware(c fiber.Ctx) error {
	if err := h.authWithGoogle(c); err != nil {
		return err
	}

	return c.Next()
}

func (h *AuthHandler) authWithGoogle(c fiber.Ctx) error {
	strToken := c.Get("Authorization")
	if strToken == "" {
		return apperrors.Unauthorized("authorization token not found")
	}
	strToken = strings.TrimPrefix(strToken, jwtAuthType+" ")

	token, err := h.Firebase.VerifyIDTokenAndCheckRevoked(c.Context(), strToken)
	if err != nil {
		if firebaseAuth.IsIDTokenExpired(err) {
			return apperrors.Unauthorized("authorization token is expired")
		}

		return apperrors.Unauthorized("invalid authorization token", err)
	}

	emailRaw, ok := token.Claims["email"].(string)
	if !ok {
		return apperrors.Unauthorized("google token: email not found")
	}

	email, ok := mtype.NewEmail(emailRaw)
	if !ok {
		return apperrors.Unauthorized("google token: email is invalid")
	}
	c.Locals("email", email)

	return nil
}

// SignUpWithGoogle godoc
//
//	@Summary		Sign up via Google
//	@Description	Creates a new user based on a Google ID token.
//	@Tags			auth
//	@Param			Authorization	header		string				true	"Authentication token (Google ID Token)"
//	@Success		200				{object}	nil					"User successfully created"
//	@Failure		401				{object}	apperrors.AppError	"Email not found in locals or invalid token"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error (e.g., user already exists)"
//	@Router			/auth/sign-up-google [post]
func (h *AuthHandler) SignUpWithGoogle(c fiber.Ctx) error {
	email, ok := c.Locals("email").(mtype.Email)
	if !ok {
		return apperrors.Unauthorized("email not found")
	}

	user, err := h.UserService.CreateWithEmail(c.Context(), email)
	if err != nil {
		return err
	}

	if user == nil {
		return apperrors.Internal("cant create user")
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

// SendCode godoc
//
//	@Summary		Send confirmation code
//	@Description	Generates and sends a one-time code to the specified email address.
//	@Tags			auth
//	@Accept			json
//	@Param			request	body		model.SendCode		true	"Send Code Request"
//	@Success		200		{object}	nil					"Code sent successfully"
//	@Failure		400		{object}	apperrors.AppError	"Bad Request or invalid email format"
//	@Failure		500		{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/auth/send-code [post]
func (h *AuthHandler) SendCode(c fiber.Ctx) error {
	var req model.SendCode
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	email, ok := mtype.NewEmail(req.Email)
	if !ok {
		return apperrors.BadRequest("invalid email")
	}

	err := h.AuthService.SendCodeOnEmail(c.Context(), "index.html", email)
	if err != nil {
		return err
	}

	return nil
}

// SignUpWithEmail godoc
//
//	@Summary		Sign up via email
//	@Description	Creates a new user by verifying an email and confirmation code.
//	@Tags			auth
//	@Accept			json
//	@Param			request	body		model.SignUpWithEmail	true	"Sign Up Request"
//	@Success		200		{object}	nil						"User successfully created"
//	@Failure		400		{object}	apperrors.AppError		"Bad Request"
//	@Failure		401		{object}	apperrors.AppError		"Invalid confirmation code"
//	@Failure		500		{object}	apperrors.AppError		"Internal Server Error (e.g., user already exists)"
//	@Router			/auth/sign-up-email [post]
func (h *AuthHandler) SignUpWithEmail(c fiber.Ctx) error {
	var req model.SignUpWithEmail
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

	user, err := h.UserService.CreateWithEmail(c.Context(), email)
	if err != nil {
		return err
	}

	if user == nil {
		return apperrors.Internal("cant create user")
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

func (h *AuthHandler) AuthMiddleware(c fiber.Ctx) error {
	token := c.Get("Authorization")
	claims, err := h.UserService.JWTAuth.ParseToken(token, 0)
	if err != nil {
		return apperrors.Unauthorized("failed to parse token", err)
	}

	if claims == nil {
		return apperrors.Internal("failed to generate claims")
	}

	c.Locals("claims", *claims)

	return c.Next()
}

func (h *AuthHandler) FindUserByEmailMiddleware(c fiber.Ctx) error {
	email, ok := c.Locals("email").(mtype.Email)
	if !ok {
		return apperrors.Unauthorized("email is missing")
	}

	user, err := h.UserService.GetByEmail(c.Context(), email)
	if err != nil {
		return apperrors.Unauthorized("user with email not found", err)
	}

	c.Locals("claims", auth.TokenClaims{
		UserID: user.ID,
	})

	return c.Next()
}
