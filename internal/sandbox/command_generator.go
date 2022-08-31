package sandbox

import (
	"github.com/alessio/shellescape"
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
)

func GenerateCommand(params *models.CommandParams) (string, error) {
	result := make(map[string]string)
	sandboxPath := "/liman/sandbox/php/index.php"

	settings := liman.GetSettings(params.User, params.Server, params.Extension)
	extension, _ := liman.GetExtension(&models.Extension{ID: params.Extension})
	server, _ := liman.GetServer(&models.Server{ID: params.Server})

	extension.Name = shellescape.StripUnsafe(extension.Name)

	if !helpers.IsLetter(extension.Name) {
		return "", fiber.NewError(fiber.StatusUnprocessableEntity, "Extension names can only contains letters")
	}

	// TODO: complete the command generator

	return "", nil
}
