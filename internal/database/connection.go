package database

import (
	"sync"

	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"gorm.io/gorm"
)

var once sync.Once
var connection *gorm.DB

func Connection() *gorm.DB {
	once.Do(func() {
		connection = initialize()
	})

	return connection
}

func initialize() *gorm.DB {
	switch helpers.Env("DB_CONNECTION", "pgsql") {
	case "pgsql":
		return initializePostgres()
	case "mysql":
		return initializeMysql()
	default:
		logger.Sugar().Fatalln("You must specify a database driver. Choices are 'postgres' or 'mysql'")
		return nil
	}
}
