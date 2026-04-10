package v1

import (
	"bufio"
	"encoding/json"
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type WalletHandler struct {
	Validator     *validator.Validate
	WalletService *service.WalletService
}

func NewWalletHandler(service *service.WalletService, validate *validator.Validate) *WalletHandler {
	return &WalletHandler{
		WalletService: service,
		Validator:     validate,
	}
}

// GenerateSolanaWallets godoc
//
//	@Summary		Generate Solana wallets
//	@Description	Generates a specified number of new Solana wallets and associates them with project IDs.
//	@Tags			wallets
//	@ID				generate-solana-wallets
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			request			body		model.GenerateWalletsReq	true	"Generation parameters"
//	@Success		201				{array}		model.Wallet				"Created"
//	@Failure		400				{object}	apperrors.AppError			"Bad Request"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/wallets/solana/generate [post]
//	@Security		BearerAuth
func (h *WalletHandler) GenerateSolanaWallets(c fiber.Ctx) error {
	var req model.GenerateWalletsReq

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

	res, err := h.WalletService.GenerateSolanaWallets(c.Context(), &req, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(res)
}

// MonitorSolanaWallets godoc
//
//	@Summary		Monitor Solana wallets
//	@Description	Starts monitoring of provided Solana wallets or updates monitoring settings.
//	@Description	This endpoint allows adding wallets to the monitoring service to track their activity.
//	@Tags			wallets
//	@ID				monitor-solana-wallets
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			request			body		model.MonitorWalletsReq		true	"Monitoring parameters"
//	@Success		200				{array}		model.MonitorWalletsResp	"OK"
//	@Failure		400				{object}	apperrors.AppError			"Bad Request"
//	@Failure		401				{object}	apperrors.AppError			"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError			"Internal Server Error"
//	@Router			/wallets/solana/monitor [post]
//	@Security		BearerAuth
func (h *WalletHandler) MonitorSolanaWallets(c fiber.Ctx) error {
	var req model.MonitorWalletsReq

	_, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("cant parse request", err)
	}

	if err := h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	res, err := h.WalletService.MonitorSolanaWallets(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(res)
}

// ImportSolanaWallets godoc
//
//	@Summary		Import Solana wallets
//	@Description	Imports existing Solana wallets using a list of private keys.
//	@Tags			wallets
//	@ID				import-solana-wallets
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Authentication token"
//	@Param			request			body		model.ImportWalletsReq	true	"Import parameters with private keys"
//	@Success		201				{array}		model.Wallet			"Created"
//	@Failure		400				{object}	apperrors.AppError		"Bad Request"
//	@Failure		401				{object}	apperrors.AppError		"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError		"Internal Server Error"
//	@Router			/wallets/solana/import [post]
//	@Security		BearerAuth
func (h *WalletHandler) ImportSolanaWallets(c fiber.Ctx) error {
	var req model.ImportWalletsReq

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

	wallets, err := h.WalletService.ImportWallets(c.Context(), &req, claims.UserID)

	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(wallets)
}

// ImportSolanaWalletsFromFile godoc
//
//	@Summary		Import Solana wallets from file
//	@Description	Imports Solana wallets from a text file (.txt) containing private keys (one key per line).
//	@Tags			wallets
//	@ID				import-solana-wallets-from-file
//	@Consumes		multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Param			file			formData	file				true	"Text file with private keys (one per line)"
//	@Param			project_ids		formData	string				true	"JSON array of project IDs to associate wallets with."
//	@Success		201				{array}		model.Wallet		"Created"
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/wallets/solana/import-file [post]
//	@Security		BearerAuth
func (h *WalletHandler) ImportSolanaWalletsFromFile(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return apperrors.BadRequest("cant get file from request", err)
	}

	projectIDs := c.FormValue("project_ids")
	if projectIDs == "" {
		return apperrors.BadRequest("project_ids is required")
	}

	var ids []uint64
	if err = json.Unmarshal([]byte(projectIDs), &ids); err != nil {
		return apperrors.BadRequest("invalid project_ids format", err)
	}

	uploadedFile, err := file.Open()
	if err != nil {
		return apperrors.BadRequest("cant open uploaded file", err)
	}
	defer uploadedFile.Close()

	var privateKeys []string
	scanner := bufio.NewScanner(uploadedFile)
	for scanner.Scan() {
		if trimmed := strings.TrimSpace(scanner.Text()); trimmed != "" {
			privateKeys = append(privateKeys, trimmed)
		}
	}

	if err = scanner.Err(); err != nil {
		return apperrors.BadRequest("cant read file content", err)
	}

	req := model.ImportWalletsReq{
		ProjectIDs:  ids,
		PrivateKeys: privateKeys,
	}

	if err = h.Validator.Struct(&req); err != nil {
		return apperrors.BadRequest("request is not valid", err)
	}

	wallets, err := h.WalletService.ImportWallets(c.Context(), &req, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(wallets)
}

// FetchPrivateKeyByID godoc
//
//	@Summary		Get private key by wallet ID
//	@Description	Fetches the private key for a given wallet ID.
//	@Tags			wallets
//	@ID				fetch-private-key-by-id
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Param			id				path		int					true	"Wallet ID"
//	@Success		200				{object}	model.PrivateKey	"OK"
//	@Failure		400				{object}	apperrors.AppError	"Bad Request (invalid public key)"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/wallets/solana/{id} [get]
//	@Security		BearerAuth
func (h *WalletHandler) FetchPrivateKeyByID(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	privateKey, err := h.WalletService.FetchPrivateKeyByWalletID(c.Context(), id, claims.UserID)

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(model.PrivateKey{
		PrivateKey: privateKey,
	})

}

func (h *WalletHandler) FetchPrivateKeysByProjectID(c fiber.Ctx) error {

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Query("projectID"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	wallets, err := h.WalletService.FetchPrivateKeysByProjectID(c.Context(), id, claims.UserID)

	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(wallets)

}

func (h *WalletHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	walletGroup := app.Group("/wallets")
	{
		solanaGroup := walletGroup.Group("/solana")
		solanaGroup.Use(auth.AuthMiddleware)
		{
			solanaGroup.Post("/generate", h.GenerateSolanaWallets)
			solanaGroup.Post("/monitor", h.MonitorSolanaWallets)
			solanaGroup.Post("/import", h.ImportSolanaWallets)
			solanaGroup.Post("/import-file", h.ImportSolanaWalletsFromFile)
			solanaGroup.Get("/:id", h.FetchPrivateKeyByID)
			solanaGroup.Get("", h.FetchPrivateKeysByProjectID)
		}
	}
}
