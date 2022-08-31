package models

// ServerModel Structure of the server obj
type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	IPAddress   string `json:"ip_address"`
	City        string `json:"city"`
	ControlPort string `json:"control_port"`
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Os          string `json:"os"`
	Enabled     string `json:"enabled"`
	KeyPort     int    `json:"key_port"`
	SharedKey   int    `json:"shared_key"`
}

func (Server) TableName() string {
	return "servers"
}
