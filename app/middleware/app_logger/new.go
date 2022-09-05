package app_logger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"go.uber.org/zap"
)

func New() fiber.Handler {
	return appLogger
}

func appLogger(c *fiber.Ctx) error {
	formData := helpers.GetFormData(c)
	logger.Sugar().WithOptions(
		zap.WithCaller(false),
	).Infow(
		"Render engine request",
		"log_id", uuid.NewString(),
		"user_id", c.Locals("user_id").(string),
		"server_id", formData["server_id"],
		"extension_id", formData["extension_id"],
		"route", c.Path(),
		"ip_address", c.IP(),
		"display", true,
		"view", formData["lmntargetFunction"],
		"request_details", formData,
	)

	return c.Next()
}
