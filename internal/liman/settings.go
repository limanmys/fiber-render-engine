package liman

import (
	"encoding/json"

	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/mervick/aes-everywhere/go/aes256"
)

// GetSettings takes needed parameters, decodes them and returns the settings that belongs to parameters
func GetSettings(user *models.User, server *models.Server, extension *models.Extension) (map[string]string, error) {
	extJson, err := GetExtensionJSON(extension)

	if err != nil {
		return nil, err
	}

	// Determine which global variables exists
	extensionKeys := make(map[string]any)
	globalVars := make(map[string]any)
	for _, setting := range extJson["database"].([]interface{}) {
		isGlobal := setting.(map[string]interface{})["global"]
		if isGlobal != nil && isGlobal.(bool) {
			globalVars[setting.(map[string]interface{})["variable"].(string)] = setting
		}

		extensionKeys[setting.(map[string]interface{})["variable"].(string)] = setting
	}

	settings := []*models.Settings{}
	results := make(map[string]string)

	decryptionKey := helpers.Env("APP_KEY", "") + user.ID + server.ID

	// Get user_settings for user and decrypt it
	searchKeys := []string{}
	for key := range extensionKeys {
		searchKeys = append(searchKeys, key)
	}
	database.Connection().Find(
		&settings,
		"name IN ? AND user_id = ? AND server_id = ?",
		searchKeys,
		user.ID,
		server.ID,
	)
	for _, setting := range settings {
		results[setting.Name] = aes256.Decrypt(setting.Value, decryptionKey)
		if extensionKeys[setting.Name].(map[string]any)["type"].(string) == "server" {
			serverObject, err := GetServer(&models.Server{ID: results[setting.Name]})
			if err != nil {
				serverObject = &models.Server{}
			}

			serverJson, _ := json.Marshal(serverObject)
			results[setting.Name] = string(serverJson)
		}
	}

	// Search global variables shared between users
	searchKeys = []string{}
	for key := range globalVars {
		searchKeys = append(searchKeys, key)
	}
	database.Connection().Find(
		&settings,
		"name IN ? AND server_id = ?",
		searchKeys,
		server.ID,
	)
	for _, setting := range settings {
		if _, ok := results[setting.Name]; ok {
			continue
		}

		results[setting.Name] = aes256.Decrypt(setting.Value, helpers.Env("APP_KEY", "")+setting.UserID+setting.ServerID)
		if extensionKeys[setting.Name].(map[string]any)["type"].(string) == "server" {
			serverObject, err := GetServer(&models.Server{ID: results[setting.Name]})
			if err != nil {
				serverObject = &models.Server{}
			}

			serverJson, _ := json.Marshal(serverObject)
			results[setting.Name] = string(serverJson)
		}
	}

	return results, nil
}
