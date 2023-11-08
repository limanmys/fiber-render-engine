package process_queue

import (
	b64 "encoding/base64"
	"strings"

	"github.com/google/uuid"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/sandbox"
	"github.com/limanmys/render-engine/internal/user_token"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/linux"
	"gorm.io/gorm"
)

type CreateReport struct {
	Queue *models.Queue
	DB    *gorm.DB
}

func (c CreateReport) Process() error {
	// Update cronjob as processing
	c.Queue.UpdateStatus(models.StatusProcessing)
	// Check extension
	extension, err := liman.GetExtension(&models.Extension{
		ID: c.Queue.Data["extension_id"].(string),
	})
	if err != nil {
		// Update job as failed
		c.Queue.UpdateError(err.Error())
		return err
	}
	// Check extension status
	if extension.Status == "0" {
		// Update job as failed
		c.Queue.UpdateError("extension is unavailable")
		return err
	}

	// Get credentials
	credentials := &models.Credentials{}
	if extension.RequireKey == "true" {
		credentials, err = liman.GetCredentials(
			&models.User{
				ID: c.Queue.Data["user_id"].(string),
			},
			&models.Server{
				ID: c.Queue.Data["server_id"].(string),
			},
		)
		// Check errors and username is valid
		if err != nil || len(credentials.Username) < 1 {
			// Update job as failed
			c.Queue.UpdateError("you need a key to use this extension")
			return err
		}
	}

	// Encode to b64 and set as form value
	formValues := make(map[string]string)
	formValues["data"] = b64.StdEncoding.EncodeToString([]byte(c.Queue.Data["payload"].(string)))

	// Generate token for user
	token, err := user_token.Create(c.Queue.Data["user_id"].(string))
	if err != nil {
		// Update job as failed
		c.Queue.UpdateError(err.Error())
		return err
	}

	// Generate new id for logs
	log_id := uuid.New()

	// Generate command
	command, err := sandbox.GenerateCommand(
		extension,
		credentials,
		&models.CommandParams{
			TargetFunction: c.Queue.Data["target"].(string),
			Locale:         helpers.Env("APP_LANG", "tr"),
			Extension:      c.Queue.Data["extension_id"].(string),
			Server:         c.Queue.Data["server_id"].(string),
			User:           c.Queue.Data["user_id"].(string),
			LogID:          log_id.String(),
			RequestData:    formValues,
			BaseURL:        "https://127.0.0.1",
			Token:          token,
		},
	)
	if err != nil {
		// Update job as failed
		c.Queue.UpdateError(err.Error())

		return err
	}

	output := linux.Execute(command)

	c.Queue.UpdateAsDone(strings.TrimSpace(strings.ReplaceAll(output, "\"", "")))
	return nil
}
