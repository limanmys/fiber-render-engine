package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

// SetExtensionDb changes specified vault keys from extension
func SetExtensionDb(c *fiber.Ctx) error {
	params := []string{"target", "new_param", "server_id", "extension_id", "token"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	extJson, err := liman.GetExtensionJSON(&models.Extension{ID: c.FormValue("extension_id")})
	if err != nil {
		return err
	}

	isGlobal, isWritable := false, false
	for _, setting := range extJson["database"].([]interface{}) {
		option := setting.(map[string]interface{})

		if option["variable"] != c.FormValue("target") {
			continue
		}

		if option["global"] != nil && option["global"].(bool) {
			isGlobal = true
		}

		if option["writable"] != nil && option["writable"].(bool) {
			isWritable = true
		}
	}

	if !isWritable {
		return c.SendString(c.FormValue("new_param"))
	}

	output, err := liman.SetExtensionDb(
		c.FormValue("new_param"),
		c.FormValue("target"),
		c.Locals("user_id").(string),
		c.FormValue("server_id"),
		isGlobal,
	)
	if err != nil {
		return err
	}

	return c.SendString(output)
}
