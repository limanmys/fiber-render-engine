package database

import (
	"fmt"

	"github.com/limanmys/render-engine/pkg/helpers"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initializeMysql() *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		helpers.Env("DB_USERNAME", ""),
		helpers.Env("DB_PASSWORD", ""),
		helpers.Env("DB_HOST", "127.0.0.1"),
		helpers.Env("DB_PORT", "3306"),
		helpers.Env("DB_DATABASE", ""),
	)

	connection, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil
	}
	return connection
}
