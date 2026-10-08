package liman

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
)

// GetPermissions Gets user and extensions permissons and variables
func GetPermissions(user *models.User, extFilter string) ([]string, map[string]interface{}, error) {
	roles, err := getRoleMaps(user)
	if err != nil {
		return nil, nil, err
	}

	permissions := []string{}
	variables := make(map[string]interface{})
	for _, role := range roles {
		permission, variable, err := getPermissionsFromMorph(role, strings.ToLower(extFilter))
		if err != nil {
			return nil, nil, err
		}

		permissions = append(permissions, permission...)

		variables = helpers.MergeInterfaceMaps(variables, variable)
	}

	if user.AuthType == "keycloak" {
		token := &models.Oauth2Token{}
		database.Connection().First(&token, "user_id = ?", user.ID)

		if token.UserID != "" {
			permissions = append(permissions, token.Permissions...)
		}
	}

	return permissions, variables, nil
}

// GetObjectPermissions Returns object permissions of user
func GetObjectPermissions(user *models.User) ([]models.Permission, error) {
	roles, err := getRoleMaps(user)
	if err != nil {
		return nil, err
	}

	permissions := []models.Permission{}
	for _, role := range roles {
		permission, err := getObjectPermissionsFromMorph(role)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission...)
	}

	return permissions, nil
}

// getPermissionsFromMorph Searches db for morph relationships and returns permissions
func getPermissionsFromMorph(morphID string, extFilter string) ([]string, map[string]interface{}, error) {
	permission := []*models.Permission{}

	err := database.Connection().Find(&permission, "morph_id = ?", morphID).Error
	if err != nil {
		return nil, nil, logger.FiberError(fiber.StatusInternalServerError, "error while fetching the permissions")
	}

	funcPerms := []string{}
	varPerms := make(map[string]interface{})

	for _, item := range permission {
		if item.Type == "function" {
			if len(extFilter) > 0 {
				if extFilter == item.Value {
					funcPerms = append(funcPerms, item.Extra)
				}
			} else {
				funcPerms = append(funcPerms, item.Extra)
			}
		}

		if item.Type == "variable" {
			if item.Extra == "multiselect" || item.Extra == "array" {
				var value []interface{}
				json.Unmarshal([]byte(item.Value), &value)
				varPerms[item.Key] = value
				continue
			}

			// Check if item is a json object
			if item.Extra == "json" {
				var value interface{}
				json.Unmarshal([]byte(item.Value), &value)
				varPerms[item.Key] = value
				continue
			}

			if existing, ok := varPerms[item.Key]; ok {
				if existingSlice, ok := existing.([]string); ok {
					existingSlice = append(existingSlice, item.Value)
					varPerms[item.Key] = existingSlice
				} else {
					// If it's not a slice, convert it to a slice
					varPerms[item.Key] = []string{existing.(string), item.Value}
				}
			} else {
				varPerms[item.Key] = item.Value
			}
		}
	}

	return funcPerms, varPerms, nil
}

// getObjectPermissionsFromMorph Searches db for morph relationships and returns object permissions
func getObjectPermissionsFromMorph(morphID string) ([]models.Permission, error) {
	permissions := []models.Permission{}

	err := database.Connection().Find(&permissions, "morph_id = ? AND type IN ? AND key = ? AND extra IS NULL", morphID, []string{"server", "extension"}, "id").Error
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// HasObjectPermission keeps object types distinct, matching core's permission contract.
func HasObjectPermission(permissions []models.Permission, objectType, id string) bool {
	if id == "" {
		return false
	}
	for _, permission := range permissions {
		if permission.Type == objectType && permission.Key == "id" && permission.Value == id && permission.Extra == "" {
			return true
		}
	}
	return false
}

// getRoleMaps Retrieves role mapping for users
func getRoleMaps(user *models.User) ([]string, error) {
	roles := []*models.RoleUsers{}

	err := database.Connection().Find(&roles, "user_id = ?", user.ID).Error
	if err != nil {
		return nil, logger.FiberError(fiber.StatusInternalServerError, "error while fetching the roles")
	}

	roleID := []string{user.ID}
	for _, item := range roles {
		roleID = append(roleID, item.RoleID)
	}

	return roleID, nil
}
