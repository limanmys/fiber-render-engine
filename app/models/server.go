package models

// ServerModel Structure of the server obj
type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	IPAddress   string `json:"ip_address"`
	ControlPort string `json:"control_port"`
	UserID      string `json:"-"`
	CreatedAt   string `json:"-"`
	UpdatedAt   string `json:"-"`
	Os          string `json:"os"`
	Enabled     string `json:"enabled"`
	KeyPort     int    `json:"key_port"`
	SharedKey   int    `json:"shared_key"`
}

func (Server) TableName() string {
	return "servers"
}
