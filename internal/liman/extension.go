package liman

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
)

func GetExtension(extension *models.Extension) (*models.Extension, error) {
	result := database.Connection().First(&extension)

	if result.Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Cannot found extension with this id")
	}

	if result.RowsAffected > 0 {
		return extension, nil
	}

	return nil, fiber.NewError(fiber.StatusNotFound, "Cannot found extension with this id")
}

func GetExtensionJSON(extension *models.Extension) (map[string]any, error) {
	fileName := ""
	if len(extension.Name) > 0 {
		fileName = strings.ToLower(extension.Name)
	} else {
		extObject, err := GetExtension(extension)
		fileName = strings.ToLower(extObject.Name)

		if err != nil {
			return nil, err
		}
	}

	jsonFile, err := ioutil.ReadFile(fmt.Sprintf("/liman/extensions/%s/db.json", fileName))

	if err != nil {
		return nil, err
	}

	extJson := make(map[string]any)
	sonic.Unmarshal(jsonFile, &extJson)

	return extJson, nil
}

func GetLicence(extension *models.Extension) (*models.Licence, error) {
	licence := &models.Licence{}

	rows := database.Connection().Where(&licence, "extension_id = ?", extension.ID).RowsAffected

	if rows > 0 {
		return licence, nil
	}

	return nil, fiber.NewError(fiber.StatusNotFound, "Licence not found")
}
