package models

// Extension Structure of the extension obj
type Extension struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Icon       string `json:"icon"`
	Service    string `json:"service"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	Order      int    `json:"order"`
	SslPorts   string `json:"sslPorts" pg:"sslPorts"`
	Issuer     string `json:"issuer"`
	Language   string `json:"language"`
	Support    string `json:"support"`
	Displays   string `json:"displays"`
	Status     string `json:"status"`
	RequireKey string `json:"require_key"`
}

func (Extension) TableName() string {
	return "extensions"
}
