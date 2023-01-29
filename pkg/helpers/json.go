package helpers

import (
	"encoding/json"
)

// Check if string is a valid JSON
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// Check if output is type of Liman JSON format
func LimanJSON(str string) bool {
	type output struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}

	var out output
	return json.Unmarshal([]byte(str), &out) == nil
}
