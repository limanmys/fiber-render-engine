package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
)

func CredentialTest(c *fiber.Ctx) error {

	credentials, err := bridge.GetCredentials(
		&models.User{
			ID: c.Params("user"),
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
