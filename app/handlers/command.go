package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

func CommandRunner(c *fiber.Ctx) error {
	if len(c.FormValue("server_id")) < 1 {
		return logger.FiberError(fiber.StatusUnprocessableEntity, "server not found")
	}

	server, err := liman.GetServer(&models.Server{ID: c.FormValue("server_id")})
	if err != nil {
		return err
	}

	shell, err := bridge.GetConnection(c.Locals("user_id").(string), c.FormValue("server_id"), server.IPAddress)
	if err != nil {
		return err
	}

	output, err := shell.Run(c.FormValue("command"))
	if err != nil {
		return logger.FiberError(fiber.StatusForbidden, "cannot run command")
	}

	return c.JSON(output)
}
