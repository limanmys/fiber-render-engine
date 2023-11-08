package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/internal/process_queue"
	"github.com/limanmys/render-engine/pkg/helpers"
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
	var formData map[string]string
	queue := &models.Queue{}
	if err := c.BodyParser(&queue); err != nil {
		formData = helpers.GetFormData(c)
		if formData == nil {
			return err
		}
	}

	if queue.Type == "" {
		queue = &models.Queue{
			Type: models.Operation(formData["type"]),
			Data: map[string]interface{}{
				"extension_id": formData["extension_id"],
				"server_id":    formData["server_id"],
				"user_id":      formData["user_id"],
				"target":       formData["target"],
				"payload":      formData["payload"],
			},
		}
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
	case models.OperationReport:
		processor = &process_queue.CreateReport{
			Queue: queue,
			DB:    h.db,
		}

		go processor.Process()
	}

	return c.JSON(queue)
}

func (h *QueueHandler) Index(c *fiber.Ctx) error {
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

	if c.FormValue("queue_type") == "" {
		return errors.New("invalid queue type")
	}

	var queues []*models.Queue
	if err := h.db.Model(&models.Queue{}).
		Where("type = ?", c.Params("queue_type")).
		Where("data->>'extension_id' ?", extension_id).
		Where("data->>'server_id' ?", server_id).
		Where("data->>'user_id' ?", user_id).Find(&queues).Error; err != nil {
		return err
	}

	return c.JSON(queues)
}

func (h *QueueHandler) Delete(c *fiber.Ctx) error {
	// Parse uuid
	uid_, err := uuid.Parse(c.Params("id"))
	if err != nil {
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

	if err := h.db.Model(&models.Queue{}).
		Where("type = ?", "report").
		Where("id = ?", uid_).
		Where("data->>'extension_id' ?", extension_id).
		Where("data->>'server_id' ?", server_id).
		Where("data->>'user_id' ?", user_id).Delete(models.Queue{}).Error; err != nil {
		return err
	}

	return c.JSON("Item deleted successfully.")
}
