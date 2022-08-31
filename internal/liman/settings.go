package liman

import (
	"os"

	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/mervick/aes-everywhere/go/aes256"
)

func GetSettings(userID string, serverID string, extensionID string) map[string]string {
	extJson, err := GetExtensionJSON(&models.Extension{ID: extensionID})

	if err != nil {
		return nil
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

	settings := []models.Settings{}
	results := make(map[string]string)

	decryptionKey := os.Getenv("APP_KEY") + userID + serverID

	// Get user_settings for user and decrypt it
	database.Connection().Find(
		&settings,
		"name IN ? AND user_id = ? AND server_id = ?",
		extensionKeys,
		userID,
		serverID,
	)
	for _, setting := range settings {
		results[setting.Name] = aes256.Decrypt(setting.Value, decryptionKey)
	}

	// Search global variables shared between users
	database.Connection().Find(
		&settings,
		"name IN ? AND server_id = ?",
		globalVars,
		serverID,
	)
	for _, setting := range settings {
		if _, ok := results[setting.Name]; ok {
			continue
		}

		results[setting.Name] = aes256.Decrypt(setting.Value, decryptionKey)
	}

	return results
}
