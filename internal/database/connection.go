package database

import (
	"sync"

	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"gorm.io/gorm"
)

var once sync.Once
var connection *gorm.DB

// Connection creates singleton pattern to connect to database only once and keep it in memory
func Connection() *gorm.DB {
	once.Do(func() {
		connection = initialize()
	})

	return connection
}

// initialize Initializes database connection
func initialize() *gorm.DB {
	switch helpers.Env("DB_CONNECTION", "pgsql") {
	case "pgsql":
		return initializePostgres()
	//case "mysql":
	//	return initializeMysql()
	default:
		//logger.Sugar().Fatalln("You must specify a database driver. Choices are 'postgres' or 'mysql'")
		logger.Sugar().Fatalln("You must specify a database driver. Choices are 'pgsql'")
		return nil
	}
}
