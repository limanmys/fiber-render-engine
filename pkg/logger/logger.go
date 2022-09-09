package logger

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func InitLogger() {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.OutputPaths = []string{
		"stdout",
		"/liman/logs/liman_new.log",
	}

	var err error
	logger, err = cfg.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
}

type Zapper interface {
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
}

func Logger() *zap.Logger {
	return logger
}

func Zap() *zap.Logger {
	return logger.WithOptions(zap.AddCallerSkip(1))
}

func Sugar() *zap.SugaredLogger {
	return logger.Sugar().WithOptions(zap.AddCallerSkip(1))
}

func FiberError(code int, message string) *fiber.Error {
	switch {
	case code >= 500:
		Sugar().Errorw(
			message,
			"lmn_level", "error",
			"code", code,
		)
	case code >= 400:
		Sugar().Warnw(
			message,
			"lmn_level", "warn",
			"code", code,
		)
	default:
		Sugar().Infow(
			message,
			"lmn_level", "info",
			"code", code,
		)
	}

	return fiber.NewError(code, message)
}
