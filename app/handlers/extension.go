package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/sandbox"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/linux"
	"github.com/limanmys/render-engine/pkg/logger"
	"go.uber.org/zap"
)

// ExtensionRunner runs extensions and returns rendered HTML/JSON views
func ExtensionRunner(c *fiber.Ctx) error {
	if len(c.FormValue("extension_id")) < 1 {
		return logger.FiberError(fiber.StatusBadRequest, "extension not found")
	}

	extension, err := liman.GetExtension(&models.Extension{
		ID: c.FormValue("extension_id"),
	})

	if err != nil {
		return err
	}

	if extension.Status == "0" {
		return logger.FiberError(fiber.StatusServiceUnavailable, "extension is unavailable")
	}

	credentials := &models.Credentials{}
	if extension.RequireKey == "true" {
		credentials, err = liman.GetCredentials(
			&models.User{
				ID: c.Locals("user_id").(string),
			},
			&models.Server{
				ID: c.FormValue("server_id"),
			},
		)

		if err != nil || len(credentials.Username) < 1 {
			return logger.FiberError(fiber.StatusForbidden, "you need a key to use this extension")
		}
	}

	formValues := helpers.GetFormData(c)

	token := c.FormValue("token")
	if len(c.FormValue("liman-token")) > 0 {
		token = c.FormValue("liman-token")
	}

	command, err := sandbox.GenerateCommand(
		extension,
		credentials,
		&models.CommandParams{
			TargetFunction: c.FormValue("lmntargetFunction"),
			User:           c.Locals("user_id").(string),
			Extension:      c.FormValue("extension_id"),
			Server:         c.FormValue("server_id"),
			RequestData:    formValues,
			Token:          token,
			BaseURL:        c.FormValue("lmnbaseurl", c.Get("origin")),
			Locale:         c.FormValue("locale", helpers.Env("APP_LANG", "tr")),
			LogID:          c.Locals("log_id").(string),
		},
	)

	if err != nil {
		return err
	}

	output := linux.Execute(command)

	if helpers.IsJSON(output) {
		type LimanMessage struct {
			Message any `json:"message"`
			Status  int `json:"status"`
		}
		msg := &LimanMessage{}

		err := json.Unmarshal([]byte(output), &msg)
		if err != nil {
			return c.Type("json").SendString(output)
		}

		if msg != nil && msg.Status > 399 {
			return c.Status(201).Type("json").SendString(output)
		}

		if msg != nil {
			return c.Status(msg.Status).Type("json").SendString(output)
		}

		return c.Type("json").SendString(output)
	}

	return c.SendString(output)
}

func ExtensionLogger(c *fiber.Ctx) error {
	params := []string{
		"title",
		"message",
	}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	formData := helpers.GetFormData(c)

	user_id := ""
	if c.Locals("user_id") != nil {
		user_id = c.Locals("user_id").(string)
	}

	logger.Sugar().WithOptions(
		zap.WithCaller(false),
	).Infow(
		"send log handler",
		"lmn_level", "high_level",
		"log_id", c.Locals("log_id").(string),
		"user_id", user_id,
		"route", "/",
		"ip_address", c.IP(),
		"request_details", formData,
	)

	return c.Type("json").SendString(`{
		"status":  200,
		"message": "log added successfully"
	}`)
}
