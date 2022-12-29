package liman

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/logger"
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

	err := json.Unmarshal(
		[]byte(serverKey.Data),
		encryptedKey,
	)
	if err != nil {
		return nil, err
	}

	credentials := encryptedKey.DecryptData(&models.User{ID: encrypterUser}, server)
	credentials.Type = serverKey.Type

	if len(credentials.Username) < 1 {
		return nil, logger.FiberError(fiber.StatusNotFound, "server not found")
	}

	return credentials, nil
}
