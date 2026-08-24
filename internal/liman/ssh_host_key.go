package liman

import (
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
)

// GetTrustedSshHostKeys returns every active key approved for an SSH endpoint.
func GetTrustedSshHostKeys(host string, port int) ([]models.SshHostKey, error) {
	keys := []models.SshHostKey{}
	result := database.Connection().
		Where("host = ? AND port = ? AND revoked_at IS NULL", host, port).
		Find(&keys)

	return keys, result.Error
}
