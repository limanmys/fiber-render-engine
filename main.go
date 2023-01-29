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

/*
Main

Initialization point of application
*/
func main() {
	// This application needs to run as root
	id := linux.Execute("id -u")
	if strings.Trim(id, "\n") != "0" {
		log.Fatalln("You need to run the service as root")
	}

	// Initialization of logger
	logger.InitLogger()
	defer logger.Logger().Sync()

	// Read environment file
	err := godotenv.Load(constants.CORE_PATH + "/.env")
	if err != nil {
		logger.Sugar().Fatalln("Cannot read Liman environment file")
	}

	// Clean long standing sessions from memory
	go bridge.Clean()

	// Start web server
	utils.CreateServer()
}
