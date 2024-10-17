package handlers

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/file"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/sandbox"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
)

// PutFile puts file to remote end from local path
func PutFile(c *fiber.Ctx) error {
	params := []string{"server_id", "remote_path", "local_path"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	_, err := file.PutFileHandler(
		c.Locals("user_id").(string),
		c.FormValue("server_id"),
		c.FormValue("remote_path"),
		c.FormValue("local_path"),
	)
	if err != nil {
		return err
	}

	return c.Type("json").SendString(`{
		"status":  200,
		"message": "file transfer completed successfully"
	}`)
}

// GetFile downloads file from remote path to local path
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

	return c.Type("json").SendString(`{
		"status":  200,
		"message": "file transfer completed successfully"
	}`)
}

// DownloadFile returns file pointer
func DownloadFile(c *fiber.Ctx) error {
	params := []string{"server_id", "path", "extension_id"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}
	if len(c.FormValue("extension_id")) < 1 {
		return logger.FiberError(fiber.StatusBadRequest, "extension not found")
	}

	extension, err := liman.GetExtension(&models.Extension{
		ID: c.FormValue("extension_id"),
	})

	if err != nil {
		return err
	}

	credentials := &models.Credentials{}
	if extension.RequireKey == "true" {
		credentials, err = liman.GetCredentials(
			&models.User{
				ID: c.Locals("user_id").(string),
			},
			&models.Server{
				ID: c.FormValue("server_id"),
			},
		)

		if err != nil || len(credentials.Username) < 1 {
			return logger.FiberError(fiber.StatusForbidden, "you need a key to use this extension")
		}
	}

	formValues := helpers.GetFormData(c)

	_, err = sandbox.GenerateCommand(
		extension,
		credentials,
		&models.CommandParams{
			TargetFunction: "",
			User:           c.Locals("user_id").(string),
			Extension:      c.FormValue("extension_id"),
			Server:         c.FormValue("server_id"),
			RequestData:    formValues,
			Token:          c.Locals("token").(string),
			BaseURL:        c.FormValue("lmnbaseurl", c.Get("origin")),
			Locale:         c.FormValue("locale", helpers.Env("APP_LANG", "tr")),
			LogID:          c.Locals("log_id").(string),
		},
	)
	if err != nil {
		return err
	}

	path := filepath.Clean(c.FormValue("path"))
	path = "/liman/extensions/" + strings.ToLower(extension.Name) + path
	path = strings.ReplaceAll(path, "../", "/")
	path = strings.ReplaceAll(path, "./", "/")
	path = strings.ReplaceAll(path, "..", "")

	return c.Download(path)
}
