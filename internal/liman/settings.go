package liman

import (
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/mervick/aes-everywhere/go/aes256"
)

func GetSettings(user *models.User, server *models.Server, extension *models.Extension) (map[string]string, error) {
	extJson, err := GetExtensionJSON(extension)

	if err != nil {
		return nil, err
	}

	// Determine which global variables exists
	var extensionKeys []string
	var globalVars []string
	for _, setting := range extJson["database"].([]interface{}) {
		isGlobal := setting.(map[string]interface{})["global"]
		if isGlobal != nil && isGlobal.(bool) {
			globalVars = append(globalVars, setting.(map[string]interface{})["variable"].(string))
		}

		extensionKeys = append(extensionKeys, setting.(map[string]interface{})["variable"].(string))
	}

	settings := []*models.Settings{}
	results := make(map[string]string)

	decryptionKey := helpers.Env("APP_KEY", "") + user.ID + server.ID

	// Get user_settings for user and decrypt it
	database.Connection().Find(
		&settings,
		"name IN ? AND user_id = ? AND server_id = ?",
		extensionKeys,
		user.ID,
		server.ID,
	)
	for _, setting := range settings {
		results[setting.Name] = aes256.Decrypt(setting.Value, decryptionKey)
	}

	// Search global variables shared between users
	database.Connection().Find(
		&settings,
		"name IN ? AND server_id = ?",
		globalVars,
		server.ID,
	)
	for _, setting := range settings {
		if _, ok := results[setting.Name]; ok {
			continue
		}

		results[setting.Name] = aes256.Decrypt(setting.Value, decryptionKey)
	}

	return results, nil
}
