package liman

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
)

func GetServer(server *models.Server) (*models.Server, error) {
	result := database.Connection().First(&server)

	if result.Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "cannot found server with this id")
	}

	if result.RowsAffected > 0 {
		return server, nil
	}

	return nil, fiber.NewError(fiber.StatusNotFound, "cannot found server with this id")
}
