package models

// UserModel Structure of the users table
type User struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Status        int    `json:"status"`
	LastLoginAt   string `json:"last_login_at"`
	RememberToken string `json:"remember_token"`
	LastLoginIP   string `json:"last_login_ip"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	ForceChange   bool   `json:"forcechange" pg:"forceChange"`
	ObjectGUID    string `json:"objectguid" pg:"objectguid"`
	AuthType      string `json:"auth_type"`
}

func (User) TableName() string {
	return "users"
}
