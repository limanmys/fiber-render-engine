package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
)

func CredentialTest(c *fiber.Ctx) error {

	credentials, err := liman.GetCredentials(
		&models.User{
			ID: c.Locals("user_id").(string),
		},
		&models.Server{
			ID: c.Params("server"),
		},
	)

	if err != nil {
		return err
	}

	return c.JSON(credentials)
}
