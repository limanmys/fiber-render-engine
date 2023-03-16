package models

type Oauth2Token struct {
	UserID           string `json:"user_id"`
	TokenType        string `json:"token_type"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

func (Oauth2Token) TableName() string {
	return "oauth2_tokens"
}
