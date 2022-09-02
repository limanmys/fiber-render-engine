package helpers

import "os"

func Env(key string, def string) string {
	if len(os.Getenv(key)) > 0 {
		return os.Getenv(key)
	}

	return def
}
