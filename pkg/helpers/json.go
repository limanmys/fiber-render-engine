package helpers

import (
	"encoding/json"

	"github.com/bytedance/sonic"
)

func IsJSON(str string) bool {
	var js json.RawMessage
	return sonic.Unmarshal([]byte(str), &js) == nil
}
