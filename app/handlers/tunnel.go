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
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	sshPort := c.FormValue("ssh_port")
	if len(sshPort) < 1 {
		sshPort = "22"
	}

	port := bridge.CreateTunnel(
		c.FormValue("remote_host"),
		c.FormValue("remote_port"),
		c.FormValue("username"),
		c.FormValue("password"),
		sshPort,
	)

	return c.JSON(port)
}

func KeepTunnelAlive(c *fiber.Ctx) error {
	params := []string{"remote_host", "remote_port", "username"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	_, err := bridge.Tunnels.Get(
		c.FormValue("remote_host"),
		c.FormValue("remote_port"),
		c.FormValue("username"),
	)
	if err != nil {
		return logger.FiberError(fiber.StatusNotFound, "tunnel not found")
	}

	return c.Type("json").SendString(`{
		"status":  200,
		"message": "tunnel keep alive successfully"
	}`)
}
