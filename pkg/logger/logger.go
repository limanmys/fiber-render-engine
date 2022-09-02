package logger

import (
	"log"

	"go.uber.org/zap"
)

var (
	Logger *zap.Logger
)

func InitLogger() {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"stdout",
		"/liman/logs/fiber.log",
	}

	var err error
	Logger, err = cfg.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
}

func Zap() *zap.Logger {
	return Logger.WithOptions(zap.AddCallerSkip(1))
}

func Sugar() *zap.SugaredLogger {
	return Logger.Sugar().WithOptions(zap.AddCallerSkip(1))
}
