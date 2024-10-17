package models

// Extension Structure of the extension obj
type Extension struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	VersionCode string `json:"version_code"`
	Icon        string `json:"-"`
	CreatedAt   string `json:"-"`
	UpdatedAt   string `json:"-"`
	SslPorts    string `json:"sslPorts" pg:"sslPorts"`
	RequireKey  string `json:"require_key"`
}

func (Extension) TableName() string {
	return "extensions"
}
