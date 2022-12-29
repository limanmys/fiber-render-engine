package handlers

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

func ScriptRunner(c *fiber.Ctx) error {
	params := []string{"local_path", "root", "parameters", "server_id"}

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
		c.FormValue("server_id"),
		server.IPAddress,
	)
	if !established {
		return logger.FiberError(fiber.StatusServiceUnavailable, "cannot establish file connection")
	}

	remotePath := ""
	if server.Os == "linux" {
		remotePath = "/tmp/" + filepath.Base(c.FormValue("local_path"))
		_, err := session.Run("rm " + remotePath)
		if err != nil {
			return err
		}
	} else {
		remotePath = session.WindowsPath + c.FormValue("remote_path") + ".ps1"
	}

	err = session.Put(c.FormValue("local_path"), remotePath)
	if err != nil {
		return err
	}

	output := ""
	if server.Os == "linux" {
		_, err := session.Run("chmod +x " + remotePath)
		if err != nil {
			return err
		}

		if c.FormValue("root") == "yes" {
			credentials, err := liman.GetCredentials(&models.User{ID: c.Locals("user_id").(string)}, server)
			if err != nil {
				return err
			}

			if credentials.Type == "ssh" {
				remotePath = `sudo -p "liman-pass-sudo" ` + remotePath
			} else {
				remotePath = "sudo " + remotePath
			}
		}

		output, err = session.Run(remotePath + " " + c.FormValue("parameters"))
		if err != nil {
			return logger.FiberError(fiber.StatusInternalServerError, "cannot run linux script")
		}
	} else {
		output, err = session.Run(session.WindowsLetter + ":\\" + remotePath + " " + c.FormValue("parameters"))
		if err != nil {
			return logger.FiberError(fiber.StatusInternalServerError, "cannot run windows script")
		}
	}

	return c.SendString(output)
}
