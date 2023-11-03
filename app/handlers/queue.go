package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/process_queue"
	"github.com/limanmys/render-engine/pkg/logger"
	"gorm.io/gorm"
)

type QueueHandler struct {
	db *gorm.DB
}

func NewQueueHandler() *QueueHandler {
	return &QueueHandler{
		db: database.Connection(),
	}
}

func (h *QueueHandler) Create(c *fiber.Ctx) error {
	queue := &models.Queue{}
	if err := c.BodyParser(&queue); err != nil {
		return err
	}

	if queue.Data["server_id"] == nil {
		return c.Status(422).JSON(fiber.Map{
			"server_id": "server id is required",
		})
	}

	// Check if user is eligible to use this server
	user, _ := liman.GetUser(&models.User{ID: c.Locals("user_id").(string)})
	if user.Status == 0 {
		var count int64
		h.db.Model(&models.Permission{}).Find(&models.Permission{
			MorphID: c.Locals("user_id").(string),
			Type:    "server",
			Value:   queue.Data["server_id"].(string),
		}).Count(&count)

		if count < 1 {
			return logger.FiberError(fiber.StatusForbidden, "you are not allowed to use this server")
		}
	}

	// Create queue object
	if err := h.db.Create(queue).Error; err != nil {
		return err
	}

	// Start processing by type
	var processor process_queue.ProcessQueue
	switch queue.Type {
	case models.OperationInstall:
		processor = &process_queue.InstallPackage{
			Queue:  queue,
			DB:     h.db,
			UserID: c.Locals("user_id").(string),
		}
		go processor.Process()
	}

	return c.JSON(queue)
}
