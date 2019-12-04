package helpers

// IsEmpty returns if form data is empty or not
func IsEmpty(data string) bool {
	if len(data) == 0 {
		return true
	} else {
		return false
	}
}
