package models

// UserModel Structure of the users table
type User struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Password      string `json:"-" gorm:"-"`
	Status        int    `json:"status"`
	LastLoginAt   string `json:"-" gorm:"-"`
	RememberToken string `json:"-" gorm:"-"`
	LastLoginIP   string `json:"-" gorm:"-"`
	CreatedAt     string `json:"-"`
	UpdatedAt     string `json:"-"`
	ForceChange   bool   `json:"-" pg:"forceChange"`
	ObjectGUID    string `json:"objectguid" pg:"objectguid"`
	AuthType      string `json:"auth_type"`
	OidcSub       string `json:"oidc_sub"`
}

func (User) TableName() string {
	return "users"
}
