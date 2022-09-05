package main

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/pkg/linux"
	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/limanmys/render-engine/pkg/utils"
)

func main() {
	id := linux.Execute("id -u")
	if strings.Trim(id, "\n") != "0" {
		log.Fatalln("You need to run the service as root")
	}

	logger.InitLogger()
	defer logger.Logger().Sync()

	err := godotenv.Load(constants.CORE_PATH + "/.env")
	if err != nil {
		logger.Sugar().Fatalln("Cannot read Liman environment file")
	}

	go bridge.Clean()

	utils.CreateServer()
}
