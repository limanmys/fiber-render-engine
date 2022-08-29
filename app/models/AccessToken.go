package models

// AccessToken Structure of the personal access token
type AccessToken struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UserID     string `json:"user_id"`
	LastUsedAt string `json:"last_used_at"`
	LastUsedIP string `json:"last_used_ip"`
	Token      string `json:"token"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	IPRange    string `json:"ip_range"`
}

func (AccessToken) TableName() string {
	return "access_tokens"
}
