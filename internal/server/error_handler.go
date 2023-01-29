package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"go.uber.org/zap"
)

/*
ErrorHandler utilizes a global GoFiber error handler and returns a Liman type error output

https://docs.gofiber.io/guide/error-handling/#custom-error-handler
*/
var ErrorHandler = func(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	var message *fiber.Map
	if code == fiber.StatusOK {
		message = &fiber.Map{
			"status":  code,
			"message": err.Error(),
		}
	} else {
		if code >= fiber.StatusInternalServerError {
			request := helpers.GetFormData(c)

			for k := range request {
				if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "token") {
					request[k] = ""
				}
			}

			logger.Sugar().WithOptions(zap.WithCaller(false)).Errorw(
				"recover middleware catch",
				"status", code,
				"message", err.Error(),
				"request_details", request,
			)
		}

		message = &fiber.Map{
			"status":  code,
			"message": err.Error(),
		}
	}

	// TODO: fix that 201 issue on new versions for liman
	// maalesef :(
	return c.Status(201).JSON(message)
}
