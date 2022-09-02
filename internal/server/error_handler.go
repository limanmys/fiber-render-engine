package server

import (
	"github.com/gofiber/fiber/v2"
)

var ErrorHandler = func(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	message := &fiber.Map{}
	if code == fiber.StatusOK {
		message = &fiber.Map{
			"message": err.Error(),
		}
	} else {
		message = &fiber.Map{
			"error": err.Error(),
		}
	}

	return c.Status(code).JSON(message)
}
