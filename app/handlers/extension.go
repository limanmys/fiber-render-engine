package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/internal/liman"
)

func ExtensionRunner(c *fiber.Ctx) error {
	/*if len(c.FormValue("extension_id")) < 1 {
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

	if extension.RequireKey == "true" {
		credentials, err := liman.GetCredentials(
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
	} */

	settings := liman.GetSettings(c.Locals("user_id").(string), c.FormValue("server_id"), c.FormValue("extension_id"))

	return c.JSON(settings)
}
