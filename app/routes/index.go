package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/limanmys/render-engine/app/handlers"
)

// Install initializes all routes of application
func Install(app *fiber.App) {
	// extension
	app.Post("/", handlers.ExtensionRunner)

	// command
	app.Post("/command", handlers.CommandRunner)
	app.Post("/outsideCommand", handlers.OutsideCommandRunner)

	// tunnel
	app.Post("/openTunnel", handlers.OpenTunnel)
	app.Post("/keepTunnelAlive", handlers.KeepTunnelAlive)

	// file
	app.Post("/getFile", handlers.GetFile)
	app.Post("/putFile", handlers.PutFile)
	app.Get("/download", handlers.DownloadFile)

	// script
	app.Post("/script", handlers.ScriptRunner)

	// verify credentials
	app.Post("/verify", handlers.Verify)

	// extensionDb
	app.Post("/setExtensionDb", handlers.SetExtensionDb)

	// logger
	app.Post("/sendLog", handlers.ExtensionLogger)

	// background job
	app.Post("/backgroundJob", handlers.BackgroundJob)

	// external api proxy
	app.Post("/externalAPI", handlers.ExternalAPI)

	// metrics
	app.Get("/metrics", monitor.New())

	// queue handler
	queueHandler := handlers.NewQueueHandler()
	app.Post("/queue", queueHandler.Create)

	// cronjob
	app.Post("/cronjobs", handlers.CreateCronJob)
	app.Get("/cronjobs", handlers.IndexCronJobs)
	app.Delete("/cronjobs/:id", handlers.DeleteCronJob)
}
