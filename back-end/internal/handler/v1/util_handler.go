package v1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

type UtilHandler struct {
	cachedIP string
	mu       sync.RWMutex
}

func NewUtilHandler() *UtilHandler {
	h := &UtilHandler{}
	go h.startIPUpdater()
	return h
}

func (h *UtilHandler) startIPUpdater() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	h.updateIP()

	for range ticker.C {
		h.updateIP()
	}
}

func (h *UtilHandler) updateIP() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip, err := getPublicIP(ctx)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.cachedIP = ip
	h.mu.Unlock()
}

func (h *UtilHandler) GetPublicIp(c fiber.Ctx) error {
	h.mu.RLock()
	ip := h.cachedIP
	h.mu.RUnlock()

	if ip == "" {
		return c.Status(http.StatusInternalServerError).
			JSON(fiber.Map{
				"error": "public IP address is not available yet",
			})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"ip": ip})
}

func (h *UtilHandler) Ping(c fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("This is a 'Market Making' service.")
}

func (h *UtilHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/", h.Ping)
	app.Get("/ip", h.GetPublicIp)
}

func getPublicIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.ipify.org", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.New("external service connection failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("external service returned non-200 status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
