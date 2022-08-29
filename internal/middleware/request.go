package middleware

import "github.com/gofiber/fiber/v2"

func RequestHandler(c *fiber.Ctx) error {
	if c.FormValue("liman-token") == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Authorization token is missing.")
	}

	

	return c.Next()
}
