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

func ExternalAPI(c *fiber.Ctx) error {
	extension := &models.Extension{}
	var err error
	if len(c.FormValue("extension_id")) > 0 {
		if !helpers.CheckUUID(c.FormValue("extension_id")) {
			extension, err = liman.GetExtension(&models.Extension{Name: c.FormValue("extension_id")})
			if err != nil {
				return err
			}
		} else {
			extension, err = liman.GetExtension(&models.Extension{ID: c.FormValue("extension_id")})
			if err != nil {
				return err
			}
		}
	} else {
		return logger.FiberError(fiber.StatusBadRequest, "extension not found")
	}

	if extension.Status == "0" {
		return logger.FiberError(fiber.StatusServiceUnavailable, "extension is unavailable")
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

	token := c.FormValue("token")
	if len(c.FormValue("liman-token")) > 0 {
		token = c.FormValue("liman-token")
	}

	command, err := sandbox.GenerateCommand(
		extension,
		credentials,
		&models.CommandParams{
			TargetFunction: c.FormValue("lmntargetFunction"),
			User:           c.Locals("user_id").(string),
			Extension:      c.FormValue("extension_id"),
			Server:         c.FormValue("server_id"),
			RequestData:    formValues,
			Token:          token,
			BaseURL:        c.FormValue("lmnbaseurl", c.Get("origin")),
			Locale:         c.FormValue("locale", helpers.Env("APP_LANG", "tr")),
		},
	)
	if err != nil {
		return err
	}

	output := linux.Execute(command)

	if helpers.IsJSON(output) {
		return c.Type("json").SendString(output)
	} else {
		return c.SendString(output)
	}
}
