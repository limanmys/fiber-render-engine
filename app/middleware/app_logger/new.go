package app_logger

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"go.uber.org/zap"
)

// Creates a new app logger instance
// This logger logs requests and removes safe information like password etc.
func New() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		c.Locals("log_id", uuid.NewString())

		log_level := helpers.Env("NEW_LOG_LEVEL", "2")
		switch log_level {
		case "0":
			if !helpers.Contains(logger.ALL, c.Path()) {
				return c.Next()
			}
		case "1":
			if !helpers.Contains(logger.MINIMAL, c.Path()) {
				return c.Next()
			}
		case "2":
			if !helpers.Contains(logger.EXT_LOG, c.Path()) {
				return c.Next()
			}
		case "3":
			if !helpers.Contains(logger.EXT_DETAIL, c.Path()) {
				return c.Next()
			}
		default:
			return c.Next()
		}

		formData := helpers.GetFormData(c)

		for k := range formData {
			if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "token") {
				formData[k] = ""
			}
		}

		user_id := ""
		if c.Locals("user_id") != nil {
			user_id = c.Locals("user_id").(string)
		}

		logger.Sugar().WithOptions(
			zap.WithCaller(false),
		).Infow(
			"render engine request",
			"lmn_level", "request",
			"log_id", c.Locals("log_id").(string),
			"user_id", user_id,
			"route", c.Path(),
			"ip_address", c.IPs(),
			"request_details", formData,
		)

		return c.Next()
	}
}
