package v1

import (
	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
)

type SwaggerHandler struct {
}

func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

func (h *SwaggerHandler) RegisterRoutes(app *fiber.App) {
	cfg := swaggerui.Config{
		BasePath: "/",
		FilePath: "./docs/api/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}

	app.Use(swaggerui.New(cfg))
}
