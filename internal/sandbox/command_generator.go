package sandbox

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/mervick/aes-everywhere/go/aes256"
)

func GenerateCommand(extension *models.Extension, credentials *models.Credentials, params *models.CommandParams) (string, error) {
	extension.Name = shellescape.StripUnsafe(extension.Name)

	if !helpers.IsLetter(extension.Name) {
		return "", logger.FiberError(fiber.StatusUnprocessableEntity, "extension names can only contains letters")
	}

	server, user, settings, err := getParams(extension, credentials, params)
	if err != nil {
		return "", err
	}

	permissions, variables, err := liman.GetPermissions(user, extension.Name)
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
	permissionsJson, _ := sonic.Marshal(permissions)
	variablesJson, _ := sonic.Marshal(variables)

	extensionData := map[string]string{
		"server":          string(serverJson),
		"extension":       string(extensionJson),
		"settings":        string(settingsJson),
		"user":            string(userJson),
		"permissions":     string(permissionsJson),
		"variables":       string(variablesJson),
		"requestData":     string(requestData),
		"publicPath":      fmt.Sprintf(constants.EXTENSION_PUBLIC_PATH, params.BaseURL, extension.ID),
		"functionsPath":   fmt.Sprintf("%s/%s%s", constants.EXTENSIONS_PATH, strings.ToLower(extension.Name), constants.FUNCTIONS_FILE_PATH),
		"navigationRoute": fmt.Sprintf(constants.NAVIGATION_ROUTE, extension.ID, server.City, server.ID),
		"key_type":        credentials.Type,
		"function":        params.TargetFunction,
		"license":         licenceData,
		"token":           params.Token,
		"locale":          params.Locale,
		"log_id":          "0000000", // TODO: add log handlers
		"ajax":            "true",
		"apiRoute":        "/extensionRun",
	}

	secureKey, err := ioutil.ReadFile(constants.KEYS_PATH + "/" + extension.ID)
	if err != nil {
		return "", logger.FiberError(fiber.StatusNotFound, "cannot found extension key file")
	}

	extensionDataJson, _ := sonic.Marshal(extensionData)
	encryptedData := aes256.Encrypt(string(extensionDataJson), string(secureKey))

	// TODO: extJsonfile
	// TODO: required param tester
	// TODO: targetFunction and permission match check
	// TODO: so file handler

	command := fmt.Sprintf(
		"runuser %s -c 'timeout %s /usr/bin/php -d display_errors=on %s %s %s'",
		strings.Replace(extension.ID, "-", "", -1),
		helpers.Env("EXTENSION_TIMEOUT", "30"),
		constants.SANDBOX_PATH,
		constants.KEYS_PATH+"/"+extension.ID,
		encryptedData,
	)

	// TODO: complete the command generator

	return command, nil
}

func getParams(
	extension *models.Extension,
	credentials *models.Credentials,
	params *models.CommandParams,
) (*models.Server, *models.User, map[string]string, error) {
	server, err := liman.GetServer(&models.Server{ID: params.Server})
	if err != nil {
		return nil, nil, nil, err
	}

	user, err := liman.GetUser(&models.User{ID: params.User})
	if err != nil {
		return nil, nil, nil, err
	}

	settings, err := liman.GetSettings(user, server, extension)
	if err != nil {
		return nil, nil, nil, err
	}

	return server, user, settings, nil
}
