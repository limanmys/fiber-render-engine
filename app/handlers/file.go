package handlers

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

func PutFile(c *fiber.Ctx) error {
	params := []string{"server_id", "remote_path", "local_path"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusUnprocessableEntity, param+" parameter is missing")
		}
	}

	server, err := liman.GetServer(&models.Server{ID: c.FormValue("server_id")})
	if err != nil {
		return err
	}

	session, err := bridge.GetSession(c.Locals("user_id").(string), c.FormValue("server_id"), server.IPAddress)
	if err != nil {
		return err
	}

	session.CreateFileConnection(c.Locals("user_id").(string), c.FormValue("server_id"), server.IPAddress)

	remotePath := ""
	if server.Os == "linux" {
		remotePath = "/tmp/" + filepath.Base(c.FormValue("remote_path"))
	} else {
		remotePath = session.WindowsPath + c.FormValue("remote_path")
	}

	return c.JSON(remotePath)
}
