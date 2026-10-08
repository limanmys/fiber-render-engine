package permission

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
)

// Creates a new permission middleware instance
func New() fiber.Handler {
	return permissionHandler(liman.GetUser, liman.GetObjectPermissions, logger.FiberError)
}

// permission Middleware checks for if users permission is enough
func permissionHandler(getUser func(*models.User) (*models.User, error), getPermissions func(*models.User) ([]models.Permission, error), reportError func(int, string) *fiber.Error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUser(&models.User{
			ID: c.Locals("user_id").(string),
		})
		if err != nil {
			return reportError(fiber.StatusInternalServerError, "error while getting the user")
		}

		if user.Status == 1 {
			return c.Next()
		}

		perms, err := getPermissions(user)
		if err != nil {
			return reportError(fiber.StatusInternalServerError, "error while getting the object permissions")
		}

		if len(c.FormValue("server_id")) > 0 {
			if !liman.HasObjectPermission(perms, "server", c.FormValue("server_id")) {
				return reportError(fiber.StatusForbidden, "you have no permission to do this")
			}
		}

		if len(c.FormValue("extension_id")) > 0 {
			extensionID := c.FormValue("extension_id")
			if !helpers.CheckUUID(extensionID) {
				extension, err := liman.GetExtension(&models.Extension{Name: extensionID})
				if err != nil {
					return err
				}

				extensionID = extension.ID
			}

			if !liman.HasObjectPermission(perms, "extension", extensionID) {
				return reportError(fiber.StatusForbidden, "you have no permission to do this")
			}

			c.Locals("extension_id", extensionID)
		}

		return c.Next()
	}
}
