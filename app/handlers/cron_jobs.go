package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/cron_jobs"
)

// CreateCronJob creates new cron job
func CreateCronJob(c *fiber.Ctx) error {
	// Parse payload
	var payload models.CronJob
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	extension_id, err := uuid.Parse(c.FormValue("extension_id"))
	if err != nil {
		return errors.New("invalid extension id")
	}

	user_id, err := uuid.Parse(c.FormValue("user_id"))
	if err != nil {
		return errors.New("invalid user id")
	}

	server_id, err := uuid.Parse(c.FormValue("server_id"))
	if err != nil {
		return errors.New("invalid server id")
	}

	// Fill default fields
	cj := models.NewCronJob()
	payload.ID = cj.ID
	payload.Status = cj.Status
	payload.Message = cj.Message

	payload.BaseURL = "https://127.0.0.1"
	payload.ExtensionID = &extension_id
	payload.ServerID = &server_id
	payload.UserID = &user_id

	// Create cronjob rule on db
	if err := database.Connection().Model(&models.CronJob{}).Create(&payload).Error; err != nil {
		return err
	}

	// Register and run cronjob
	if err := cron_jobs.RegisterAndRun(&payload); err != nil {
		return err
	}

	return c.JSON("Cronjob registered successfully.")
}

// IndexCronJobs lists all cron jobs
func IndexCronJobs(c *fiber.Ctx) error {
	// Set empty variable for later use
	var cronjobs []*models.CronJob
	// Get all cronjobs
	if err := database.Connection().Model(&models.CronJob{}).Find(&cronjobs).Error; err != nil {
		return err
	}

	return c.JSON(cronjobs)
}

// DeleteCronJob deletes specified cron job
func DeleteCronJob(c *fiber.Ctx) error {
	// Parse uuid
	uid_, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return err
	}

	// Remove cronjob from global scheduler
	if err := cron_jobs.Delete(&uid_); err != nil {
		return err
	}

	// If cronjob successfully remove by scheduler, remove it from storage
	if err := database.Connection().Model(&models.CronJob{}).
		Where("id = ?", uid_).Delete(&models.CronJob{}).Error; err != nil {
		return err
	}

	return c.JSON("Item deleted successfully.")
}
