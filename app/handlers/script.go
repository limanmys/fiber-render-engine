package handlers

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/file"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

// ScriptRunner runs scripts on remote server
func ScriptRunner(c *fiber.Ctx) error {
	params := []string{"local_path", "root", "parameters", "server_id"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	if c.FormValue("root") != "yes" && c.FormValue("root") != "no" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid root option")
	}
	localFile, err := openExtensionFile(c, "scripts", false)
	if err != nil {
		return err
	}
	defer localFile.Close()

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

	name := filepath.Base(c.FormValue("local_path"))
	if server.Os != "linux" {
		name = c.FormValue("remote_path", name)
		if !strings.HasSuffix(strings.ToLower(name), ".ps1") {
			name += ".ps1"
		}
	}
	remotePath, err := file.RemotePath(server.Os, session.WindowsPath, name)
	if err != nil {
		return err
	}
	executionPath := remotePath
	if server.Os != "linux" {
		executionPath = session.WindowsLetter + ":\\" + remotePath
	}
	command, err := file.ScriptCommand(executionPath, c.FormValue("parameters"), server.Os != "linux")
	if err != nil {
		return err
	}
	if server.Os == "linux" {
		stagedPath, cleanup, err := session.StageLinuxScript(localFile, filepath.Base(remotePath))
		if err != nil {
			return err
		}
		defer cleanup()
		command, err = file.ScriptCommand(stagedPath, c.FormValue("parameters"), false)
		if err != nil {
			return err
		}
	} else {
		if err := session.Put(localFile, remotePath); err != nil {
			return err
		}
	}

	output := ""
	if server.Os == "linux" {
		if c.FormValue("root") == "yes" {
			credentials, err := liman.GetCredentials(&models.User{ID: c.Locals("user_id").(string)}, server)
			if err != nil {
				return err
			}

			if credentials.Type == "ssh" {
				command = `sudo -p "liman-pass-sudo" ` + command
			} else {
				command = "sudo " + command
			}
		}

		output, err = session.Run(command)
		if err != nil {
			return logger.FiberError(fiber.StatusInternalServerError, "cannot run linux script")
		}
	} else {
		output, err = session.Run(command)
		if err != nil {
			return logger.FiberError(fiber.StatusInternalServerError, "cannot run windows script")
		}
	}

	return c.SendString(output)
}
