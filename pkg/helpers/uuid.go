package helpers

import "github.com/google/uuid"

func CheckUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
