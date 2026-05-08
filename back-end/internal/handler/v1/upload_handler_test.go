package v1

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"mm/config"
	"mm/internal/client/lighthouse"
	"mm/internal/model"
	auth "mm/pkg/jwt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testLighthouseAPIKey = "yourApiKey"

func newUploadTestApp(t *testing.T) *fiber.App {
	t.Helper()

	cfg := &config.Config{
		Lighthouse: config.LighthouseConfig{
			ApiKey: testLighthouseAPIKey,
		},
	}

	h := NewUploadHandler(lighthouse.NewClient(cfg))

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	fakeAuth := func(c fiber.Ctx) error {
		sessionID, _ := uuid.NewV7()
		c.Locals("claims", auth.TokenClaims{
			UserID:    1,
			SessionID: sessionID,
		})
		return c.Next()
	}

	upload := app.Group("/upload")
	upload.Use(fakeAuth)
	upload.Post("/image", h.uploadImage)
	upload.Post("/metadata", h.uploadMetadata)

	return app
}

// smallPNG returns a 2×2 red PNG image encoded as bytes.
func smallPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	red := color.RGBA{R: 255, A: 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, red)
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestUploadImage_Success(t *testing.T) {
	app := newUploadTestApp(t)

	imgData := smallPNG(t)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("image", "test.png")
	require.NoError(t, err)
	_, err = part.Write(imgData)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload/image", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := app.Test(req, fiber.TestConfig{Timeout: -1})
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", raw)

	var result model.UploadImageResponse
	require.NoError(t, json.Unmarshal(raw, &result))

	assert.NotEmpty(t, result.CID, "CID must not be empty")
	assert.NotEmpty(t, result.URL, "URL must not be empty")
	assert.Contains(t, result.URL, result.CID, "URL must contain the CID")
	assert.Equal(t, "test.png", result.Name)
	t.Logf("uploaded image CID: %s", result.CID)
	t.Logf("gateway URL:        %s", result.URL)
}

func TestUploadImage_MissingField(t *testing.T) {
	app := newUploadTestApp(t)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "test.png")
	require.NoError(t, err)
	_, err = part.Write(smallPNG(t))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload/image", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := app.Test(req, fiber.TestConfig{Timeout: -1})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUploadMetadata_Success(t *testing.T) {
	app := newUploadTestApp(t)

	metadata := map[string]any{
		"name":        "Test Token",
		"symbol":      "TT",
		"description": "Integration test token metadata",
		"image":       "https://gateway.lighthouse.storage/ipfs/bafkreiexample",
		"attributes": []map[string]any{
			{"trait_type": "rarity", "value": "common"},
		},
	}

	raw, err := json.Marshal(metadata)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/upload/metadata", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: -1})
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", respBody)

	var result model.UploadMetadataResponse
	require.NoError(t, json.Unmarshal(respBody, &result))

	assert.NotEmpty(t, result.CID, "CID must not be empty")
	assert.NotEmpty(t, result.MetadataURL, "MetadataURL must not be empty")
	assert.Contains(t, result.MetadataURL, result.CID, "MetadataURL must contain the CID")
	t.Logf("uploaded metadata CID: %s", result.CID)
	t.Logf("gateway URL:           %s", result.MetadataURL)
}

func TestUploadMetadata_EmptyBody(t *testing.T) {
	app := newUploadTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/upload/metadata", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: -1})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUploadMetadata_InvalidJSON(t *testing.T) {
	app := newUploadTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/upload/metadata", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: -1})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
