package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

// Creates a new authorization instance
func New() fiber.Handler {
	return authorization
}

// authorization Middleware auths users before requests
func authorization(c *fiber.Ctx) error {
	if len(c.FormValue("token")) > 0 {
		user, err := liman.AuthWithToken(
			strings.Trim(c.FormValue("token"), ""),
		)

		if err != nil {
			return logger.FiberError(fiber.StatusUnauthorized, err.Error())
		}

		c.Locals("user_id", user)
		return c.Next()
	}

	if len(c.FormValue("liman-token")) > 0 {
		user, err := liman.AuthWithAccessToken(
			strings.Trim(c.FormValue("liman-token"), ""),
		)

		if err != nil {
			return logger.FiberError(fiber.StatusUnauthorized, err.Error())
		}

		c.Locals("user_id", user)
		return c.Next()
	}

	return logger.FiberError(fiber.StatusUnauthorized, "authorization token is missing")
}
