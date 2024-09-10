package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Oauth2Token struct {
	UserID           string   `json:"user_id"`
	TokenType        string   `json:"token_type"`
	AccessToken      string   `json:"access_token"`
	RefreshToken     string   `json:"refresh_token"`
	ExpiresIn        int      `json:"expires_in"`
	RefreshExpiresIn int      `json:"refresh_expires_in"`
	Permissions      StrArray `json:"permissions" gorm:"type:jsonb;index,type:gin"`
}

func (Oauth2Token) TableName() string {
	return "oauth2_tokens"
}

type StrArray []string

// Value Marshal
func (a StrArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan Unmarshal
func (a *StrArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
