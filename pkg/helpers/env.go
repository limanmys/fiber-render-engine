package helpers

import "os"

// Env function utilizes support for default values on os.Getenv
func Env(key string, def string) string {
	if len(os.Getenv(key)) > 0 {
		return os.Getenv(key)
	}

	return def
}
