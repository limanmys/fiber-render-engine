package database

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestPostgresSQLDBRejectsInvalidConnection(t *testing.T) {
	connection := &gorm.DB{Config: &gorm.Config{}}

	db, err := postgresSQLDB(connection)
	if !errors.Is(err, gorm.ErrInvalidDB) {
		t.Fatalf("postgresSQLDB() error = %v, want %v", err, gorm.ErrInvalidDB)
	}
	if db != nil {
		t.Fatal("postgresSQLDB() returned a database for an invalid connection")
	}
}
