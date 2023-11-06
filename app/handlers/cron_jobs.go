package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/cron_jobs"
)

func Create(c *fiber.Ctx) error {
	// Parse payload
	var payload models.CronJob
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

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
