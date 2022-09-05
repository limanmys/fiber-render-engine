package utils

import (
	"log"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/limanmys/render-engine/app/middleware/app_logger"
	"github.com/limanmys/render-engine/app/middleware/auth"
	"github.com/limanmys/render-engine/app/middleware/permission"
	"github.com/limanmys/render-engine/app/routes"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/internal/server"
)

func CreateServer() {
	// Create Fiber App
	app := fiber.New(fiber.Config{
		AppName:      "Liman Render Engine",
		ServerHeader: "divergent",
		Prefork:      false,
		ErrorHandler: server.ErrorHandler,
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
	})

	// app.Use(recover.New())
	app.Use(compress.New())
	app.Use(auth.New())
	app.Use(permission.New())
	app.Use(app_logger.New())

	// Mount routes
	routes.Install(app)

	// Start server
	log.Fatal(app.ListenTLS(":2806", constants.CERT_PATH+"/liman.crt", constants.CERT_PATH+"/liman.key"))
}
