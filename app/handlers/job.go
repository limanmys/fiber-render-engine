package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/sandbox"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/linux"
	"github.com/limanmys/render-engine/pkg/logger"
)

// BackgroundJob creates a background shell command
func BackgroundJob(c *fiber.Ctx) error {
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

	command, err := sandbox.GenerateCommand(
		extension,
		credentials,
		&models.CommandParams{
			TargetFunction: c.FormValue("lmntargetFunction"),
			User:           c.Locals("user_id").(string),
			Extension:      c.FormValue("extension_id"),
			Server:         c.FormValue("server_id"),
			RequestData:    formValues,
			Token:          c.Locals("token").(string),
			BaseURL:        c.FormValue("lmnbaseurl", c.Get("origin")),
			Locale:         c.FormValue("locale", helpers.Env("APP_LANG", "tr")),
		},
	)
	if err != nil {
		return err
	}

	go linux.Execute(command)

	return c.Type("json").SendString(`{
		"status":  200,
		"message": "job dispatched successfully"
	}`)
}
