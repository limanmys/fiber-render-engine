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
)

// ExternalAPI used for running other extension functions from a different extension
func ExternalAPI(c *fiber.Ctx) error {
	var extension *models.Extension
	var err error
	if len(c.FormValue("extension_id")) > 0 {
		if !helpers.CheckUUID(c.FormValue("extension_id")) {
			extension, err = liman.GetExtension(&models.Extension{Name: c.FormValue("extension_id")})
			if err != nil {
				return err
			}
		} else {
			extension, err = liman.GetExtension(&models.Extension{ID: c.FormValue("extension_id")})
			if err != nil {
				return err
			}
		}
	} else {
		return logger.FiberError(fiber.StatusBadRequest, "extension not found")
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

	command, err := sandbox.GenerateCommand(
		extension,
		credentials,
		&models.CommandParams{
			TargetFunction: c.FormValue("lmntargetFunction"),
			User:           c.Locals("user_id").(string),
			Extension:      c.FormValue("extension_id"),
			Server:         c.FormValue("server_id"),
			RequestData:    formValues,
			Token:          c.Locals("token").(string),
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
