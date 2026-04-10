package server

import (
	"mm/config"
	"mm/internal/handler/middleware"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

func NewServer(c *config.Config, l *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ProxyHeader: "X-Forwarded-For",
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	corsConfig := cors.Config{
		AllowOrigins: strings.Split(c.HTTP.AllowOrigins, ","),
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-Forwarded-For", "X-CSRF-Token",
			"Authorization", "User-Env", "Access-Control-Request-Headers", "Access-Control-Allow-Headers",
			"Access-Control-Request-Method", "Content-Unique-Identifier", "Content-Index",
			"Access-Control-Allow-BaseError", "Access-Control-Request-Headers", "Access-Control-Request-Method",
		},
		AllowCredentials: c.HTTP.AllowCredentials,
		MaxAge:           int(12 * time.Hour),
	}

	app.Use(cors.New(corsConfig))

	logMiddleware := middleware.NewLoggingMiddleware(l)
	logMiddleware.RegisterLogger(c, app)
	app.Use(recoverer.New())

	return app
}
