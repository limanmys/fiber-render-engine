package utils

import (
	"fmt"
	"log"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/limanmys/render-engine/app/middleware/auth"
	"github.com/limanmys/render-engine/app/routes"
	"github.com/limanmys/render-engine/internal/server"
)

func CreateServer(port int) {
	// Create Fiber App
	app := fiber.New(fiber.Config{
		AppName:      "Liman Render Engine",
		Prefork:      false,
		ErrorHandler: server.ErrorHandler,
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
	})

	app.Use(recover.New())
	app.Use(compress.New())
	app.Use(auth.New())

	// Mount routes
	routes.Install(app)

	// Start server
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}