package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/handlers"
)

func Install(app *fiber.App) {
	app.Post("/extensionRunner", handlers.ExtensionRunner)
	app.Post("/commandRunner", handlers.CommandRunner)
}
