package liman

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
)

// GetPermissions Gets user and extensions permissons and variables
func GetPermissions(user *models.User, extFilter string) ([]string, map[string]string, error) {
	roles, err := getRoleMaps(user)
	if err != nil {
		return nil, nil, err
	}

	permissions := []string{}
	variables := make(map[string]string)
	for _, role := range roles {
		permission, variable, err := getPermissionsFromMorph(role, strings.ToLower(extFilter))
		if err != nil {
			return nil, nil, err
		}

		permissions = append(permissions, permission...)

		variables = helpers.MergeStringMaps(variables, variable)
	}

	return permissions, variables, nil
}

// GetObjectPermissions Returns object permissions of user
func GetObjectPermissions(user *models.User) ([]string, error) {
	roles, err := getRoleMaps(user)
	if err != nil {
		return nil, err
	}

	permissions := []string{}
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
func getPermissionsFromMorph(morphID string, extFilter string) ([]string, map[string]string, error) {
	permission := []*models.Permission{}

	err := database.Connection().Find(&permission, "morph_id = ?", morphID).Error
	if err != nil {
		return nil, nil, logger.FiberError(fiber.StatusInternalServerError, "error while fetching the permissions")
	}

	funcPerms := []string{}
	varPerms := make(map[string]string)

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
			varPerms[item.Key] = item.Value
		}
	}

	return funcPerms, varPerms, nil
}

// getObjectPermissionsFromMorph Searches db for morph relationships and returns object permissions
func getObjectPermissionsFromMorph(morphID string) ([]string, error) {
	permissions := []*models.Permission{}

	err := database.Connection().Find(&permissions, "morph_id = ? AND NOT type = 'function'", morphID).Error
	if err != nil {
		return nil, err
	}

	results := []string{}
	for _, permission := range permissions {
		results = append(results, permission.Value)
	}

	return results, nil
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
