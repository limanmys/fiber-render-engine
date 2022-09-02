package database

import (
	"fmt"

	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initializePostgres() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=Europe/Istanbul",
		helpers.Env("DB_HOST", "127.0.0.1"),
		helpers.Env("DB_PORT", "5432"),
		helpers.Env("DB_USERNAME", ""),
		helpers.Env("DB_PASSWORD", ""),
		helpers.Env("DB_DATABASE", ""),
	)

	connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logger.Sugar().Fatalln("Cannot connect to Liman database!")
	}

	db, _ := connection.DB()

	err = db.Ping()
	if err != nil {
		logger.Sugar().Fatalln("Cannot connect to Liman database!")
	}

	return connection
}
