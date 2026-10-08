package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/file"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
)

func openExtensionFile(c *fiber.Ctx, directory string, writable bool) (*os.File, error) {
	id := c.FormValue("extension_id")
	if id == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "extension_id parameter is missing")
	}
	target := &models.Extension{ID: id}
	if !helpers.CheckUUID(id) {
		target = &models.Extension{Name: id}
	}
	extension, err := liman.GetExtension(target)
	if err != nil {
		return nil, err
	}
	return file.LocalFile(extension.Name, c.FormValue("local_path"), directory, writable)
}
