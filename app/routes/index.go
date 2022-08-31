package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/handlers"
)

func Install(app *fiber.App) {
	app.Get("/credentials/:server", handlers.CredentialTest)
	app.Post("/extensionRunner", handlers.ExtensionRunner)
}
