package sandbox

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/mervick/aes-everywhere/go/aes256"
)

func GenerateCommand(extension *models.Extension, credentials *models.Credentials, params *models.CommandParams) (string, error) {
	extension.Name = shellescape.StripUnsafe(extension.Name)

	if !helpers.IsLetter(extension.Name) {
		return "", fiber.NewError(fiber.StatusUnprocessableEntity, "Extension names can only contains letters")
	}

	server, err := liman.GetServer(&models.Server{ID: params.Server})
	if err != nil {
		return "", err
	}

	user, err := liman.GetUser(&models.User{ID: params.User})
	if err != nil {
		return "", err
	}

	settings, err := liman.GetSettings(user, server, extension)
	if err != nil {
		return "", err
	}

	if credentials.Username != "" && credentials.Key != "" {
		settings["clientUsername"] = credentials.Username
		settings["clientPassword"] = credentials.Key
	}

	licence, err := liman.GetLicence(extension)
	licenceData := ""
	if err == nil {
		licenceData = licence.Data
	}

	serverJson, _ := sonic.Marshal(server)
	extensionJson, _ := sonic.Marshal(extension)
	settingsJson, _ := sonic.Marshal(settings)
	userJson, _ := sonic.Marshal(user)
	requestData, _ := sonic.Marshal(params.RequestData)

	extensionData := map[string]string{
		"server":          string(serverJson),
		"extension":       string(extensionJson),
		"settings":        string(settingsJson),
		"user":            string(userJson),
		"key_type":        credentials.Type,
		"functionsPath":   fmt.Sprintf("/liman/extensions/%s/views/functions.php", strings.ToLower(extension.Name)),
		"function":        params.TargetFunction,
		"requestData":     string(requestData),
		"license":         licenceData,
		"apiRoute":        "/extensionRun",
		"navigationRoute": fmt.Sprintf("/l/%s/%s/%s", extension.ID, server.City, server.ID),
		"token":           params.Token,
		"locale":          params.Locale,
		"log_id":          "0000000", // TODO: add log handlers
		"ajax":            "true",
		"publicPath":      fmt.Sprintf("%s/eklenti/%s/public/", params.BaseURL, extension.ID),
		"permissions":     "[]",
		"variables":       "[]",
		// TODO: Add role system
	}

	secureKey, err := ioutil.ReadFile("/liman/keys/" + extension.ID)
	if err != nil {
		return "", fiber.NewError(fiber.StatusNotFound, "Cannot found extension key file")
	}

	extensionDataJson, _ := sonic.Marshal(extensionData)
	encryptedData := aes256.Encrypt(string(extensionDataJson), string(secureKey))

	timeout := "30"
	if len(os.Getenv("EXTENSION_TIMEOUT")) > 0 {
		timeout = os.Getenv("EXTENSION_TIMEOUT")
	}

	command := fmt.Sprintf(
		"runuser %s -c 'timeout %s /usr/bin/php -d display_errors=on %s %s %s'",
		strings.Replace(extension.ID, "-", "", -1),
		timeout,
		"/liman/sandbox/php/index.php",
		"/liman/keys/"+extension.ID,
		encryptedData,
	)

	// TODO: complete the command generator

	return command, nil
}
