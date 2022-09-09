package utils

import (
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"
	"github.com/limanmys/render-engine/app/middleware/app_logger"
	"github.com/limanmys/render-engine/app/middleware/auth"
	"github.com/limanmys/render-engine/app/middleware/permission"
	"github.com/limanmys/render-engine/app/routes"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/internal/server"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/linux"
	"github.com/limanmys/render-engine/pkg/logger"
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

	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(compress.New())
	app.Use(auth.New())
	app.Use(permission.New())
	app.Use(app_logger.New())

	// Mount routes
	routes.Install(app)

	// Start server
	err := app.ListenTLS("127.0.0.1:2806", constants.CERT_PATH+"/liman.crt", constants.CERT_PATH+"/liman.key")
	if err != nil {
		logger.Sugar().Errorw("app initialization error", "details", err)
		if strings.Contains(err.Error(), "listen tcp4 :2806: bind: address already in use") {
			logger.Sugar().Infow("restarting app to freeup port")
			linux.Execute("fuser -k 2806/tcp")
			time.Sleep(time.Second)
			helpers.RestartSelf()
		}
	}
}
