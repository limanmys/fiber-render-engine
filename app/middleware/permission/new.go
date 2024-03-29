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
	return permission
}

// permission Middleware checks for if users permission is enough
func permission(c *fiber.Ctx) error {
	user, err := liman.GetUser(&models.User{
		ID: c.Locals("user_id").(string),
	})
	if err != nil {
		return logger.FiberError(fiber.StatusInternalServerError, "error while getting the user")
	}

	if helpers.Env("LIMAN_RESTRICTED", "false") == "true" {
		return c.Next()
	}

	if user.Status == 1 {
		return c.Next()
	}

	perms, err := liman.GetObjectPermissions(user)
	if err != nil {
		return logger.FiberError(fiber.StatusInternalServerError, "error while getting the object permissions")
	}

	if len(c.FormValue("server_id")) > 0 {
		if !helpers.Contains(perms, c.FormValue("server_id")) {
			return logger.FiberError(fiber.StatusForbidden, "you have no permission to do this")
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

		if !helpers.Contains(perms, extensionID) || len(extensionID) < 1 {
			return logger.FiberError(fiber.StatusForbidden, "you have no permission to do this")
		}

		c.Locals("extension_id", extensionID)
	}

	return c.Next()
}
