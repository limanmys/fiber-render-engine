package helpers

// Contains checks if string contains needle
func Contains(items []string, str string) bool {
	for _, item := range items {
		if item == str {
			return true
		}
	}

	return false
}
