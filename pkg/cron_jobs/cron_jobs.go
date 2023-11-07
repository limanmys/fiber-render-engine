package cron_jobs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/sandbox"
	"github.com/limanmys/render-engine/internal/user_token"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/linux"
)

func RegisterAndRun(cj *models.CronJob) error {
	_, err := constants.GLOBAL_SCHEDULER.Tag(cj.ID.String()).Every(1).Week().Weekday(time.Weekday(cj.Day)).At(cj.Time).Do(func() {
		// Update cronjob as processing
		cj.UpdateAsProcessing()
		// Get extension
		extension, err := liman.GetExtension(&models.Extension{
			ID: cj.ExtensionID.String(),
		})
		if err != nil {
			// Update job as failed
			cj.UpdateAsFailed(err.Error())
			return
		}
		// Check extension status
		if extension.Status == "0" {
			// Update job as failed
			cj.UpdateAsFailed("extension is unavailable")
			return
		}

		// Get credentials
		credentials := &models.Credentials{}
		if extension.RequireKey == "true" {
			credentials, err = liman.GetCredentials(
				&models.User{
					ID: cj.UserID.String(),
				},
				&models.Server{
					ID: cj.ServerID.String(),
				},
			)
			// Check errors and username is valid
			if err != nil || len(credentials.Username) < 1 {
				// Update job as failed
				cj.UpdateAsFailed("you need a key to use this extension")
				return
			}
		}

		// Parse payload
		marshalledPayload, err := json.Marshal(cj.Payload)
		if err != nil {
			cj.UpdateAsFailed(err.Error())
			return
		}
		// Set form values as map
		formValues := make(map[string]string)
		formValues["data"] = string(marshalledPayload)

		// Generate token for user
		token, err := user_token.Create(cj.UserID.String())
		if err != nil {
			cj.UpdateAsFailed(err.Error())
			return
		}

		// Generate new id for logs
		log_id := uuid.New()

		// Generate command
		command, err := sandbox.GenerateCommand(
			extension,
			credentials,
			&models.CommandParams{
				TargetFunction: cj.Target,
				Locale:         helpers.Env("APP_LANG", "tr"),
				Extension:      cj.ExtensionID.String(),
				Server:         cj.ServerID.String(),
				User:           cj.UserID.String(),
				LogID:          log_id.String(),
				RequestData:    formValues,
				BaseURL:        cj.BaseURL,
				Token:          token,
			},
		)
		if err != nil {
			// Update job as failed
			cj.UpdateAsFailed(err.Error())

			return
		}

		linux.Execute(command)
		cj.UpdateAsDone()
	})
	if err != nil {
		return err
	}

	constants.GLOBAL_SCHEDULER.StartAsync()

	return nil
}

func Delete(id *uuid.UUID) error {
	// Remove cronjob from global scheduler
	if err := constants.GLOBAL_SCHEDULER.RemoveByTag(id.String()); err != nil {
		return err
	}

	return nil
}

func InitCronJobs() error {
	var cronjobs []*models.CronJob
	if err := database.Connection().Find(&cronjobs).Error; err != nil {
		return err
	}

	for _, cronjob := range cronjobs {
		if err := RegisterAndRun(cronjob); err != nil {
			cronjob.UpdateAsFailed(err.Error())
		}
	}

	return nil
}
