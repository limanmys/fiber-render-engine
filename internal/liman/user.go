package liman

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
)

func GetUser(user *models.User) (*models.User, error) {
	result := database.Connection().First(&user)

	if result.Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Cannot found user with this id")
	}

	if result.RowsAffected > 0 {
		return user, nil
	}

	return nil, fiber.NewError(fiber.StatusNotFound, "Cannot found user with this id")
}
