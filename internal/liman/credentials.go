package liman

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/logger"
	"gorm.io/gorm"
)

type serverKeyRepository interface {
	FindUserKey(userID, serverID string) (*models.ServerKey, error)
	FindSharedKey(serverID string) (*models.ServerKey, error)
}

type gormServerKeyRepository struct {
	db *gorm.DB
}

func (r gormServerKeyRepository) FindUserKey(userID, serverID string) (*models.ServerKey, error) {
	serverKey := &models.ServerKey{}
	err := r.db.Where("user_id = ? AND server_id = ?", userID, serverID).
		Order("updated_at DESC").
		Order("id ASC").
		First(serverKey).Error

	return serverKey, err
}

func (r gormServerKeyRepository) FindSharedKey(serverID string) (*models.ServerKey, error) {
	serverKey := &models.ServerKey{}
	err := r.db.Where("server_id = ? AND shared = ?", serverID, true).
		Order("updated_at DESC").
		Order("id ASC").
		First(serverKey).Error

	return serverKey, err
}

func resolveServerKey(repository serverKeyRepository, userID, serverID string) (*models.ServerKey, error) {
	serverKey, err := repository.FindUserKey(userID, serverID)
	if err == nil {
		return serverKey, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return repository.FindSharedKey(serverID)
}

// GetCredentials Searches db and returns credentials of server
func GetCredentials(user *models.User, server *models.Server) (*models.Credentials, error) {
	serverKey, err := resolveServerKey(
		gormServerKeyRepository{db: database.Connection()},
		user.ID,
		server.ID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, logger.FiberError(fiber.StatusNotFound, "server key not found")
		}

		return nil, err
	}

	encryptedKey := &models.KeyData{}
	err = json.Unmarshal(
		[]byte(serverKey.Data),
		encryptedKey,
	)
	if err != nil {
		return nil, err
	}

	credentials := encryptedKey.DecryptData(&models.User{ID: serverKey.UserID}, server)
	credentials.Type = serverKey.Type

	if len(credentials.Username) < 1 {
		return nil, logger.FiberError(fiber.StatusNotFound, "server not found")
	}

	return credentials, nil
}
