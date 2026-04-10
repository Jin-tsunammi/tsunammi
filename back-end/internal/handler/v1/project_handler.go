package v1

import (
	"fmt"
	"mm/internal/model"
	"mm/internal/service"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"net/http"
	"slices"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type ProjectHandler struct {
	Validator      *validator.Validate
	ProjectService *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService, validate *validator.Validate) *ProjectHandler {
	return &ProjectHandler{
		ProjectService: service,
		Validator:      validate,
	}
}

// GetProjects godoc
//
//	@Summary		List projects
//	@Description	Returns all projects
//	@Tags			projects
//	@ID				get-projects
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Success		200				{array}		model.ProjectWithWalletsResponse
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/projects [get]
//	@Security		BearerAuth
func (h *ProjectHandler) GetProjects(c fiber.Ctx) error {

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

	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")

	if !slices.Contains([]string{"id", ""}, sortBy) {
		return apperrors.BadRequest("sortBy must be either 'id', 'last_sync' or 'balance'")
	}

	if !slices.Contains([]string{"asc", "desc", ""}, sortOrder) {
		return apperrors.BadRequest("sortOrder must be either 'asc' or 'desc'")
	}

	if sortBy == "" {
		sortBy = "id"
	}

	if sortOrder == "" {
		sortOrder = "asc"
	}

	res, err := h.ProjectService.FetchProjectsWithWalletsWithoutBalance(c.Context(), claims.UserID, parsedPage, parsedPageSize, sortBy, sortOrder == "desc")
	if err != nil {
		return apperrors.Internal("cant get projects", err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

// GetProjectByID godoc
//
//	@Summary		Get project by ID
//	@Description	Returns a project by its ID
//	@Tags			projects
//	@ID				get-project-by-id
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			id				path		int		true	"Project ID"
//	@Success		200				{object}	model.ProjectWithWalletsResponse
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/projects/{id} [get]
//	@Security		BearerAuth
func (h *ProjectHandler) GetProjectByID(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	res, err := h.ProjectService.FetchProjectWithWalletsByID(c.Context(), id, claims.UserID)
	if err != nil {
		return apperrors.Internal("cant get project", err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *ProjectHandler) GetProjectsWithMintedBalance(c fiber.Ctx) error {
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

	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")

	if !slices.Contains([]string{"id", ""}, sortBy) {
		return apperrors.BadRequest("sortBy must be either 'id', 'last_sync' or 'balance'")
	}

	if !slices.Contains([]string{"asc", "desc", ""}, sortOrder) {
		return apperrors.BadRequest("sortOrder must be either 'asc' or 'desc'")
	}

	if sortBy == "" {
		sortBy = "id"
	}

	if sortOrder == "" {
		sortOrder = "asc"
	}

	mint := c.Query("mint")
	if mint == "" {
		return apperrors.BadRequest("mint is required")
	}

	parsedMint, err := solana.PublicKeyFromBase58(mint)
	if err != nil {
		return apperrors.BadRequest("invalid mint", err)
	}

	res, err := h.ProjectService.FetchProjectWithWalletsByMint(c.Context(), claims.UserID, parsedPage, parsedPageSize, sortBy, sortOrder == "desc", parsedMint)
	if err != nil {
		return apperrors.Internal("cant get projects", err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *ProjectHandler) GetProjectByIDWithMintedBalance(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	mint := c.Query("mint")
	if mint == "" {
		return apperrors.BadRequest("mint is required")
	}

	parsedMint, err := solana.PublicKeyFromBase58(mint)
	if err != nil {
		return apperrors.BadRequest("invalid mint", err)
	}

	res, err := h.ProjectService.FetchProjectWithWalletByIDAndMint(c.Context(), id, claims.UserID, parsedMint)
	if err != nil {
		return apperrors.Internal("cant get projects", err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *ProjectHandler) GetCachedProjectByID(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	res, err := h.ProjectService.FetchCachedProjectWithWalletsByID(c.Context(), id, claims.UserID)
	if err != nil {
		return apperrors.Internal("cant get project", err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *ProjectHandler) GetProjectsWithoutWallets(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	fmt.Println(claims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	projects, err := h.ProjectService.FetchProjects(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(projects)
}

// CreateProject godoc
//
//	@Summary		Create project
//	@Description	Creates a new project
//	@Tags			projects
//	@ID				create-project
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Authentication token"
//	@Param			request			body		model.CreateProjectReq	true	"Create Project Request"
//	@Success		201				{object}	model.Project
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/projects [post]
//	@Security		BearerAuth
func (h *ProjectHandler) CreateProject(c fiber.Ctx) error {
	var req model.CreateProjectReq

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

	res, err := h.ProjectService.CreateProject(c.Context(), req, claims.UserID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(res)
}

// DeleteProject godoc
//
//	@Summary		Delete project
//	@Description	Deletes a project by its ID
//	@Tags			projects
//	@ID				delete-project
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication token"
//	@Param			id				path		int					true	"Project ID"
//	@Success		204				{object}	nil					"No Content"
//	@Failure		400				{object}	apperrors.AppError	"Bad Request"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		404				{object}	apperrors.AppError	"Not Found"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/projects/{id} [delete]
//	@Security		BearerAuth
func (h *ProjectHandler) DeleteProject(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	err = h.ProjectService.DeleteProject(c.Context(), id, claims.UserID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}

func (h *ProjectHandler) EditProject(c fiber.Ctx) error {

	var req model.EditProjectReq

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

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apperrors.BadRequest("id is not valid", err)
	}

	err = h.ProjectService.EditProfile(c.Context(), id, claims.UserID, req)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}

func (h *ProjectHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {

	app.Get("/debug/routes", func(c fiber.Ctx) error {
		return c.JSON(app.Stack())
	})

	projectGroup := app.Group("/projects")

	projectGroup.Use(auth.AuthMiddleware)
	{
		projectGroup.Get("", h.GetProjects)
		projectGroup.Get("/:id", h.GetProjectByID)
		projectGroup.Post("", h.CreateProject)
		projectGroup.Put("/:id", h.EditProject)
		projectGroup.Delete("/:id", h.DeleteProject)

	}

	cacheGroup := app.Group("/cache")
	cacheGroup.Use(auth.AuthMiddleware)
	cacheGroup.Get("/projects/:id", h.GetCachedProjectByID)

	mintGroup := app.Group("/mint-balance")
	mintGroup.Use(auth.AuthMiddleware)
	mintGroup.Get("/projects", h.GetProjectsWithMintedBalance)
	mintGroup.Get("/projects/:id", h.GetProjectByIDWithMintedBalance)

	withoutWalletsGroup := app.Group("/without-wallets")
	withoutWalletsGroup.Use(auth.AuthMiddleware)

	withoutWalletsGroup.Get("/projects", h.GetProjectsWithoutWallets)
}
