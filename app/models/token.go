package models

// TokenModel Structure of the tokens tableD
type Token struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (Token) TableName() string {
	return "tokens"
}
