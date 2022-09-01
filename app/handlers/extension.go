package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/sandbox"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/linux"
)

func ExtensionRunner(c *fiber.Ctx) error {
	if len(c.FormValue("extension_id")) < 1 {
		return fiber.NewError(fiber.StatusNotFound, "Extension not found")
	}

	extension, err := liman.GetExtension(&models.Extension{
		ID: c.FormValue("extension_id"),
	})

	if err != nil {
		return err
	}

	if extension.Status == "0" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Extension is unavailable right now, please try again later.")
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
			return fiber.NewError(fiber.StatusForbidden, "You need a key to use this extension, please add it through the case.")
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
			BaseURL:        c.FormValue("lmnbaseurl", c.BaseURL()),
			Locale:         c.FormValue("locale", os.Getenv("APP_LANG")),
		},
	)

	if err != nil {
		return err
	}

	output := linux.Execute(command)

	return c.SendString(output)
}
