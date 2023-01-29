package liman

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/logger"
)

// AuthWithToken Auth user with proprietary Liman token
func AuthWithToken(token string) (string, error) {
	tokenObj := &models.Token{}

	err := database.Connection().First(&tokenObj, "token = ?", token).Error

	if err != nil || len(tokenObj.UserID) < 1 {
		return "", logger.FiberError(fiber.StatusUnauthorized, "authorization token is not valid")
	}

	return tokenObj.UserID, nil
}

// AuthWithAccessToken Auth user with generated access token
func AuthWithAccessToken(token string) (string, error) {
	tokenObj := &models.AccessToken{}

	err := database.Connection().First(&tokenObj, "token = ?", token).Error

	if err != nil || len(tokenObj.UserID) < 1 {
		return "", logger.FiberError(fiber.StatusUnauthorized, "authorization token is not valid")
	}

	return tokenObj.UserID, nil
}
