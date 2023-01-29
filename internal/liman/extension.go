package liman

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/mervick/aes-everywhere/go/aes256"
)

// GetExtension searches db and returns matching extension
func GetExtension(extension *models.Extension) (*models.Extension, error) {
	result := database.Connection().Where(&extension).First(&extension)

	if result.Error != nil {
		return nil, logger.FiberError(fiber.StatusNotFound, "cannot found extension with this id")
	}

	if result.RowsAffected > 0 {
		return extension, nil
	}

	return nil, logger.FiberError(fiber.StatusNotFound, "cannot found extension with this id")
}

// GetExtensionJSON Retrieves extension's JSON file from filesystem and returns it as object
func GetExtensionJSON(extension *models.Extension) (map[string]any, error) {
	fileName := ""
	if len(extension.Name) > 0 {
		fileName = strings.ToLower(extension.Name)
	} else {
		extObject, err := GetExtension(extension)
		fileName = strings.ToLower(extObject.Name)

		if err != nil {
			return nil, err
		}
	}

	jsonFile, err := os.ReadFile(fmt.Sprintf("%s/%s/db.json", constants.EXTENSIONS_PATH, fileName))

	if err != nil {
		return nil, err
	}

	extJson := make(map[string]any)
	err = json.Unmarshal(jsonFile, &extJson)
	if err != nil {
		return nil, err
	}

	return extJson, nil
}

// GetLicence searches db and returns matching licence
func GetLicence(extension *models.Extension) (*models.Licence, error) {
	licence := &models.Licence{}

	rows := database.Connection().Find(&licence, "extension_id = ?", extension.ID).RowsAffected

	if rows > 0 {
		return licence, nil
	}

	return nil, errors.New("licence not found")
}

// SetExtensionDb sets the value of vault key
func SetExtensionDb(value, target, userID, serverID string, isGlobal bool) (string, error) {
	found := false
	settings := []*models.Settings{}

	if isGlobal {
		err := database.Connection().Where(&settings, "name = ? AND server_id = ?", target, serverID).Error
		if err != nil {
			return "", err
		}
	} else {
		err := database.Connection().Where(&settings, "name = ? AND server_id = ? AND user_id = ?", target, serverID, userID).Error
		if err != nil {
			return "", err
		}
	}

	for _, setting := range settings {
		key := helpers.Env("APP_KEY", "")

		found = true

		err := database.Connection().Model(&setting).Updates(map[string]any{
			"value":      aes256.Encrypt(value, key),
			"updated_at": time.Now().Format(time.RFC3339),
		}).Error

		if err != nil {
			return "", err
		}
	}

	if !found {
		setting := &models.Settings{
			ServerID: serverID,
			UserID:   userID,
			Name:     target,
			Value:    aes256.Encrypt(value, helpers.Env("APP_KEY", "")+userID+serverID),
		}

		err := database.Connection().Create(setting).Error
		if err != nil {
			return "", err
		}
	}

	return value, nil
}
