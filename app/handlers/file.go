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
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	server, err := liman.GetServer(&models.Server{ID: c.FormValue("server_id")})
	if err != nil {
		return err
	}

	session, err := bridge.GetSession(
		c.Locals("user_id").(string),
		server.ID,
		server.IPAddress,
	)
	if err != nil {
		return err
	}

	established := session.CreateFileConnection(
		c.Locals("user_id").(string),
		server.ID,
		server.IPAddress,
	)
	if !established {
		return logger.FiberError(fiber.StatusServiceUnavailable, "cannot establish file connection")
	}

	remotePath := ""
	if server.Os == "linux" {
		remotePath = "/tmp/" + filepath.Base(c.FormValue("remote_path"))
	} else {
		remotePath = session.WindowsPath + c.FormValue("remote_path")
	}

	err = session.Put(c.FormValue("local_path"), remotePath)
	if err != nil {
		return err
	}

	return c.JSON(&fiber.Map{
		"status":  "ok",
		"message": "file transfer completed successfully",
	})
}

func GetFile(c *fiber.Ctx) error {
	params := []string{"server_id", "remote_path", "local_path"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	server, err := liman.GetServer(&models.Server{ID: c.FormValue("server_id")})
	if err != nil {
		return err
	}

	session, err := bridge.GetSession(
		c.Locals("user_id").(string),
		server.ID,
		server.IPAddress,
	)
	if err != nil {
		return err
	}

	established := session.CreateFileConnection(
		c.Locals("user_id").(string),
		server.ID,
		server.IPAddress,
	)
	if !established {
		return logger.FiberError(fiber.StatusServiceUnavailable, "cannot establish file connection")
	}

	remotePath := ""
	if server.Os == "linux" {
		remotePath = "/tmp/" + filepath.Base(c.FormValue("remote_path"))
	} else {
		remotePath = session.WindowsPath + c.FormValue("remote_path")
	}

	err = session.Get(c.FormValue("local_path"), remotePath)
	if err != nil {
		return err
	}

	return c.JSON(&fiber.Map{
		"status":  "ok",
		"message": "file transfer completed successfully",
	})
}
