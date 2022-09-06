package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/pkg/logger"
)

func OpenTunnel(c *fiber.Ctx) error {
	params := []string{"remote_host", "remote_port", "username", "password"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusUnprocessableEntity, param+" parameter is missing")
		}
	}

	port := bridge.CreateTunnel(c.FormValue("remote_host"), c.FormValue("remote_port"), c.FormValue("username"), c.FormValue("password"))

	return c.JSON(port)
}
