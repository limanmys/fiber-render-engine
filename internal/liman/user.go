package liman

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/logger"
)

// GetUser searches db and returns matching user
func GetUser(user *models.User) (*models.User, error) {
	result := database.Connection().Where(&user).First(&user)

	if result.Error != nil {
		return nil, logger.FiberError(fiber.StatusNotFound, "cannot found user with this id")
	}

	if result.RowsAffected > 0 {
		return user, nil
	}

	return nil, logger.FiberError(fiber.StatusNotFound, "cannot found user with this id")
}
