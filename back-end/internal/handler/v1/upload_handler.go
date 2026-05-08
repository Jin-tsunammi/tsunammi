package v1

import (
	"bytes"
	"encoding/json"
	"mm/internal/client/lighthouse"
	"mm/internal/model"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"

	"github.com/gofiber/fiber/v3"
)

type UploadHandler struct {
	lh *lighthouse.Client
}

func NewUploadHandler(lh *lighthouse.Client) *UploadHandler {
	return &UploadHandler{lh: lh}
}

func (h *UploadHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	uploadGroup := app.Group("/upload")
	uploadGroup.Use(auth.AuthMiddleware)
	{
		uploadGroup.Post("/image", h.uploadImage)
		uploadGroup.Post("/metadata", h.uploadMetadata)
	}
}

// uploadImage godoc
//
//	@Summary		Upload image to Lighthouse
//	@Description	Uploads an image file to decentralized storage via Lighthouse and returns its IPFS CID and gateway URL
//	@Tags			upload
//	@ID				upload-image
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication token"
//	@Param			image			formData	file	true	"Image file to upload"
//	@Success		200				{object}	model.UploadImageResponse
//	@Failure		400				{object}	apperrors.AppError	"Bad Request — image field missing"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/upload/image [post]
//	@Security		BearerAuth
func (h *UploadHandler) uploadImage(c fiber.Ctx) error {
	_, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	file, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "image field required")
	}

	src, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to open file")
	}
	defer src.Close()

	upload, err := h.lh.UploadReader(c.Context(), file.Filename, file.Size, src)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "lighthouse upload failed: "+err.Error())
	}

	return c.JSON(model.NewUploadImageResponse(upload.Hash, upload.Name, upload.Size))
}

// uploadMetadata godoc
//
//	@Summary		Upload metadata to Lighthouse
//	@Description	Serializes an arbitrary JSON object and uploads it as metadata.json to decentralized storage via Lighthouse, returning its IPFS CID and gateway URL
//	@Tags			upload
//	@ID				upload-metadata
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Authentication token"
//	@Param			body			body		model.UploadMetadataRequest	true	"Arbitrary key-value metadata object (must not be empty)"
//	@Success		200				{object}	model.UploadMetadataResponse
//	@Failure		400				{object}	apperrors.AppError	"Bad Request — invalid or empty body"
//	@Failure		401				{object}	apperrors.AppError	"Unauthorized"
//	@Failure		500				{object}	apperrors.AppError	"Internal Server Error"
//	@Router			/upload/metadata [post]
//	@Security		BearerAuth
func (h *UploadHandler) uploadMetadata(c fiber.Ctx) error {
	_, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.UploadMetadataRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body: "+err.Error())
	}
	if len(req) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "metadata cannot be empty")
	}

	raw, err := json.Marshal(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to encode metadata")
	}

	upload, err := h.lh.UploadReader(
		c.Context(),
		"metadata.json",
		int64(len(raw)),
		bytes.NewReader(raw),
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "lighthouse upload failed: "+err.Error())
	}

	return c.JSON(model.NewUploadMetadataResponse(upload.Hash))
}
