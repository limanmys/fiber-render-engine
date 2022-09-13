package helpers

import (
	"encoding/json"

	"github.com/bytedance/sonic"
)

func IsJSON(str string) bool {
	var js json.RawMessage
	return sonic.UnmarshalString(str, &js) == nil
}

func LimanJSON(str string) bool {
	type output struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}

	var out output
	return sonic.UnmarshalString(str, &out) == nil
}
