package user_token

import (
	"time"

	"github.com/google/uuid"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	gorandom "github.com/zekiahmetbayar/go-random"
)

// Create a new token or retrieve old one
func Create(user_id string) (string, error) {
	// Search token on database
	var token models.Token
	database.Connection().Model(&models.Token{}).Where("user_id = ?", user_id).First(&token)

	// If token does not exists, create token
	if token.ID == "" {
		// Create new id for token
		uid := uuid.New()
		// Generate token
		token := generate()
		// Create token on database
		if err := database.Connection().Model(&models.Token{}).Create(models.Token{
			ID:        uid.String(),
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
			UserID:    user_id,
			Token:     token,
		}).Error; err != nil {
			return "", err
		}

		return token, nil
	}
	// Get token update date
	updateDate, err := time.Parse(time.RFC3339, token.UpdatedAt)
	if err != nil {
		return "", err
	}
	// If token updated after 6 hours
	if time.Since(updateDate).Hours() > 6 {
		// TODO: Update token
		token_str := generate()
		if err := database.Connection().Model(&token).Update("token", token_str).Error; err != nil {
			return "", err
		}
		return token_str, nil
	}

	return token.Token, nil
}

// Generate a new token
func generate() string {
	token, err := gorandom.String(false, true, false, 32)
	if err != nil {
		return ""
	}

	return token
}
