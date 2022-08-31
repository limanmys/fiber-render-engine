package models

type Credentials struct {
	// Type
	Type string `json:"type"`

	// Username
	Username string `json:"username"`

	// Certificate or password
	Key string `json:"key"`

	// Connection port
	Port string `json:"port"`
}
