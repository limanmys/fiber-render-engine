package main

import (
	"github.com/joho/godotenv"
	"github.com/limanmys/render-engine/internal/constants"
	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/limanmys/render-engine/pkg/utils"
)

func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()

	err := godotenv.Load(constants.CORE_PATH + "/.env")
	if err != nil {
		logger.Sugar().Fatalln("Cannot read Liman environment file")
	}

	utils.CreateServer()
}
