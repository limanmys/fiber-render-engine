package bridge

import (
	"encoding/json"

	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
)

func GetCredentials(user *models.User, server *models.Server) (*models.Credentials, error) {
	serverKey := &models.ServerKey{}

	database.Connection().First(&serverKey, "user_id = ? AND server_id = ?", user.ID, server.ID)

	encryptedKey := &models.KeyData{}
	encrypterUser := user.ID
	if serverKey.Data == "" {
		database.Connection().First(&server, "id = ?", server.ID)

		if server.SharedKey == 1 {
			database.Connection().First(&serverKey, "server_id = ?", server.ID)
			encrypterUser = serverKey.UserID
		}
	}

	json.Unmarshal([]byte(serverKey.Data), encryptedKey)
	credentials := encryptedKey.DecryptKey(&models.User{ID: encrypterUser}, server)
	credentials.Type = serverKey.Type

	return credentials, nil
}
