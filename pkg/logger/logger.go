package logger

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

/*
Init Logger

Creates log file if does not exist and configures Zap logger
*/
func InitLogger() {
	// TODO: Implement lumberjack log roller
	// https://gist.github.com/rnyrnyrny/a6dc926ae11951b753ecd66c00695397

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

// A Logger provides fast, leveled, structured logging. All methods are safe for concurrent use.
func Zap() *zap.Logger {
	return logger.WithOptions(zap.AddCallerSkip(1))
}

// Sugar wraps the Logger to provide a more ergonomic, but slightly slower, API. Sugaring a Logger is quite inexpensive,
// so it's reasonable for a single application to use both Loggers and SugaredLoggers,
// converting between them on the boundaries of performance-sensitive code.
func Sugar() *zap.SugaredLogger {
	return logger.Sugar().WithOptions(zap.AddCallerSkip(1))
}

// This method wraps Zap sugared logger with fiber errors
func FiberError(code int, message string) *fiber.Error {
	switch {
	case code >= 500:
		Sugar().Errorw(
			message,
			"code", code,
		)
	case code >= 400:
		Sugar().Warnw(
			message,
			"code", code,
		)
	default:
		Sugar().Infow(
			message,
			"code", code,
		)
	}

	return fiber.NewError(code, message)
}
