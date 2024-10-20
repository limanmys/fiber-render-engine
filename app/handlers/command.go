package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

// CommandRunner runs command on specified server
func CommandRunner(c *fiber.Ctx) error {
	if len(c.FormValue("server_id")) < 1 {
		return logger.FiberError(fiber.StatusBadRequest, "server not found")
	}

	server, err := liman.GetServer(&models.Server{ID: c.FormValue("server_id")})
	if err != nil {
		return err
	}

	shell, err := bridge.GetSession(
		c.Locals("user_id").(string),
		server.ID,
		server.IPAddress,
	)
	if err != nil {
		return err
	}

	output, err := shell.Run(c.FormValue("command"))
	if err != nil {
		return logger.FiberError(fiber.StatusForbidden, "cannot run command")
	}

	return c.SendString(output)
}

// OutsideCommandRunner runs command on external endpoints
func OutsideCommandRunner(c *fiber.Ctx) error {
	params := []string{
		"command",
		"connection_type",
		"remote_host",
		"remote_port",
		"username",
		"password",
		"disconnect",
	}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	session, err := bridge.Sessions.GetRaw(
		c.Locals("user_id").(string),
		c.FormValue("remote_host"),
		c.FormValue("username"),
	)
	if err != nil {
		session = &bridge.Session{}
		isCreated := session.CreateRaw(
			c.FormValue("connection_type"),
			c.FormValue("username"),
			c.FormValue("password"),
			c.FormValue("remote_host"),
			c.FormValue("remote_port"),
		)

		if isCreated {
			bridge.Sessions.SetRaw(
				c.Locals("user_id").(string),
				c.FormValue("remote_host"),
				c.FormValue("username"),
				session,
			)
		} else {
			return logger.FiberError(fiber.StatusInternalServerError, "cannot create connection")
		}
	}

	output, err := session.Run(c.FormValue("command"))
	if err != nil {
		return logger.FiberError(fiber.StatusForbidden, "cannot run command")
	}

	if c.FormValue("disconnect") == "1" {
		session.CloseAllConnections()
		bridge.Sessions.Delete(c.Locals("user_id").(string) + c.FormValue("remote_host") + c.FormValue("username"))
	}

	return c.SendString(output)
}
