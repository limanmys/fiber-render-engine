package liman

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/logger"
)

// GetServer searches db and returns matching server
func GetServer(server *models.Server) (*models.Server, error) {
	result := database.Connection().Where(&server).First(&server)

	if result.Error != nil {
		return nil, logger.FiberError(fiber.StatusNotFound, "cannot found server with this id")
	}

	if result.RowsAffected > 0 {
		return server, nil
	}

	return nil, logger.FiberError(fiber.StatusNotFound, "cannot found server with this id")
}
