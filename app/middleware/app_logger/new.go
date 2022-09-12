package app_logger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"go.uber.org/zap"
)

func New() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		formData := helpers.GetFormData(c)

		user_id := ""
		if c.Locals("user_id") != nil {
			user_id = c.Locals("user_id").(string)
		}

		logger.Sugar().WithOptions(
			zap.WithCaller(false),
		).Infow(
			"render engine request",
			"lmn_level", "request",
			"log_id", uuid.NewString(),
			"user_id", user_id,
			"route", c.Path(),
			"ip_address", c.IP(),
			"request_details", formData,
		)

		return c.Next()
	}
}
