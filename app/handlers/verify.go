package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/pkg/logger"
)

func Verify(c *fiber.Ctx) error {
	params := []string{"ip_address", "username", "password", "port", "keyType"}

	for _, param := range params {
		if len(c.FormValue(param)) < 1 {
			return logger.FiberError(fiber.StatusBadRequest, param+" parameter is missing")
		}
	}

	flag := bridge.VerifyAuth(
		c.FormValue("username"),
		c.FormValue("password"),
		c.FormValue("ip_address"),
		c.FormValue("port"),
		c.FormValue("key_type"),
	)

	if flag {
		return c.JSON("ok")
	}

	// TODO: change the way request handles on core
	c.SendString("nok")
	return c.SendStatus(201)
}
