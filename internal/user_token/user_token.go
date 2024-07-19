package user_token

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Create a new token for user purpose of internal use
func Create(user_id string) (string, error) {
	// Define the JWT claims
	claims := jwt.MapClaims{
		"sub": user_id,
		"exp": time.Now().Add(time.Minute * 15).Unix(), // Token expiration time
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	// Replace "your-secret-key" with your actual secret key
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
