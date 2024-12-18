package models

// CommandParams structure of CommandParams object
type CommandParams struct {
	TargetFunction string
	User           string
	Extension      string
	Server         string
	RequestData    map[string]string
	Token          string
	BaseURL        string
	Locale         string
	LogID          string
}
