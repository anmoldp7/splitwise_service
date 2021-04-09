package utils

func ValidateRequest(args ...string) bool {
	for _, value := range args {
		if value == "" {
			return false
		}
	}
	return true
}
