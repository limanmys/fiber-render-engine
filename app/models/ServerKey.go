package models

import (
	"os"

	"github.com/mervick/aes-everywhere/go/aes256"
)

// ServerKey Structure of the server keys
type ServerKey struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Data      string `json:"data"`
	ServerID  string `json:"server_id"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type KeyData struct {
	ClientUsername string `json:"clientUsername"`
	ClientPassword string `json:"clientPassword"`
	KeyPort        string `json:"key_port"`
}

func (ServerKey) TableName() string {
	return "server_keys"
}

func (d KeyData) DecryptKey(user *User, server *Server) *Credentials {
	key := os.Getenv("APP_KEY") + user.ID + server.ID

	return &Credentials{
		Username: aes256.Decrypt(d.ClientUsername, key),
		Key:      aes256.Decrypt(d.ClientPassword, key),
		Port:     d.KeyPort,
	}
}
