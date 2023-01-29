package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/mervick/aes-everywhere/go/aes256"
)

// GenerateCommand creates a command for running Sandbox PHP instance with required parameters
func GenerateCommand(extension *models.Extension, credentials *models.Credentials, params *models.CommandParams) (string, error) {
	// Get needed datas
	extension.Name = shellescape.StripUnsafe(extension.Name)

	if !helpers.IsLetter(extension.Name) {
		return "", logger.FiberError(fiber.StatusBadRequest, "extension names can only contains letters")
	}

	server, user, settings, err := GetParams(extension, credentials, params)
	if err != nil {
		return "", err
	}

	permissions, variables, err := liman.GetPermissions(user, extension.Name)
	if err != nil {
		return "", err
	}

	extJson, err := liman.GetExtensionJSON(extension)
	if err != nil {
		return "", err
	}

	requiredList := []string{}
	if extJson["functions"] != nil {
		for _, function := range extJson["functions"].([]interface{}) {
			fn := function.(map[string]any)
			requiredList = append(requiredList, fn["name"].(string))
		}
	}

	if user.Status != 1 && !helpers.Contains(permissions, params.TargetFunction) && helpers.Contains(requiredList, params.TargetFunction) {
		return "", logger.FiberError(fiber.StatusForbidden, "you have no permission to do this")
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

	// Convert required datas to JSON
	serverJson, _ := json.Marshal(server)
	extensionJson, _ := json.Marshal(extension)
	settingsJson, _ := json.Marshal(settings)
	userJson, _ := json.Marshal(user)
	requestData, _ := json.Marshal(params.RequestData)
	permissionsJson, _ := json.Marshal(permissions)
	variablesJson, _ := json.Marshal(variables)

	// Construct data that needed to be used on PHP Sandbox
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
		"navigationRoute": fmt.Sprintf(constants.NAVIGATION_ROUTE, extension.ID, server.ID),
		"key_type":        credentials.Type,
		"function":        params.TargetFunction,
		"license":         licenceData,
		"token":           params.Token,
		"locale":          params.Locale,
		"log_id":          params.LogID,
		"ajax":            "true",
		"apiRoute":        "/extensionRun",
	}

	secureKey, err := os.ReadFile(constants.KEYS_PATH + "/" + extension.ID)
	if err != nil {
		return "", logger.FiberError(fiber.StatusNotFound, "cannot found extension key file")
	}

	extensionDataJson, _ := json.Marshal(extensionData)
	encryptedData := aes256.Encrypt(string(extensionDataJson), string(secureKey))

	// If exist, use Liman licence system on extension runner
	soPath := "/liman/extensions/" + strings.ToLower(extension.Name) + "/liman.so"
	soCommand := ""
	if _, err := os.Stat(soPath); err == nil {
		soCommand = "-dextension=" + shellescape.Quote(soPath) + " "
	}

	// Create command
	command := ""
	if helpers.Env("CONTAINER_MODE", "false") != "true" {
		command = fmt.Sprintf(
			"runuser %s -c 'timeout %s /usr/bin/php %s -d display_errors=on %s %s %s'",
			strings.Replace(extension.ID, "-", "", -1),
			helpers.Env("EXTENSION_TIMEOUT", "30"),
			soCommand,
			constants.SANDBOX_PATH,
			constants.KEYS_PATH+"/"+extension.ID,
			encryptedData,
		)
	} else {
		command = fmt.Sprintf(
			"runuser extuser -c 'timeout %s /usr/bin/php %s -d display_errors=on %s %s %s'",
			helpers.Env("EXTENSION_TIMEOUT", "30"),
			soCommand,
			constants.SANDBOX_PATH,
			constants.KEYS_PATH+"/"+extension.ID,
			encryptedData,
		)
	}

	return command, nil
}

// Get needed parameters and return them
func GetParams(
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
