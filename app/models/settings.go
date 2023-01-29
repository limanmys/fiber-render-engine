package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Settings structure of Settings object
type Settings struct {
	ID        string `json:"id"`
	ServerID  string `json:"server_id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (Settings) TableName() string {
	return "user_settings"
}

func (settings *Settings) BeforeCreate(tx *gorm.DB) error {
	settings.ID = uuid.NewString()
	settings.CreatedAt = time.Now().Format(time.RFC3339)
	settings.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}
