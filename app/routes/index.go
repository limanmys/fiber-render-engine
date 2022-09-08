package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/handlers"
)

func Install(app *fiber.App) {
	// extension
	app.Post("/run", handlers.ExtensionRunner)

	// command
	app.Post("/runCommand", handlers.CommandRunner)
	app.Post("/runOutsideCommand", handlers.OutsideCommandRunner)

	// tunnel
	app.Post("/openTunnel", handlers.OpenTunnel)
	app.Post("/keepTunnelAlive", handlers.KeepTunnelAlive)

	// file
	app.Post("/getFile", handlers.GetFile)
	app.Post("/putFile", handlers.PutFile)

	// script
	app.Post("/runScript", handlers.ScriptRunner)
}
