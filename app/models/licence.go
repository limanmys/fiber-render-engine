package models

// Licence structure of Licence object
type Licence struct {
	ID          string `json:"id"`
	Data        string `json:"data"`
	ExtensionID string `json:"extension_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (Licence) TableName() string {
	return "licenses"
}
