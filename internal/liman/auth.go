package liman

import (
	"errors"

	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
)

func AuthWithToken(token string) (string, error) {
	tokenObj := &models.Token{}

	err := database.Connection().First(&tokenObj, "token = ?", token).Error

	if err != nil || len(tokenObj.UserID) < 1 {
		return "", errors.New("Authorization token is not valid.")
	}

	return tokenObj.UserID, nil
}

func AuthWithAccessToken(token string) (string, error) {
	tokenObj := &models.AccessToken{}

	err := database.Connection().First(&tokenObj, "token = ?", token).Error

	if err != nil || len(tokenObj.UserID) < 1 {
		return "", errors.New("Authorization token is not valid.")
	}

	return tokenObj.UserID, nil
}
