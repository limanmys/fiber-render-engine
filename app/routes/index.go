package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/controllers"
)

func Install(app *fiber.App) {
	app.Get("/credentials/:user/:server", controllers.CredentialTest)
}
