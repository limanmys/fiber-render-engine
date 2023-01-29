package helpers

import "github.com/google/uuid"

// Check if UUID is valid
func CheckUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
