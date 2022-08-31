package database

import (
	"log"
	"os"
	"sync"

	"gorm.io/gorm"
)

var once sync.Once
var connection *gorm.DB

func Connection() *gorm.DB {
	once.Do(func() {
		connection = initialize()
	})

	return connection.Debug()
}

func initialize() *gorm.DB {
	switch os.Getenv("DB_CONNECTION") {
	case "pgsql":
		return initializePostgres()
	case "mysql":
		return initializeMysql()
	default:
		log.Fatalln("You must specify a database driver. Choices are 'postgres' or 'mysql'")
		return nil
	}
}
