package utils

import "regexp"

func IsValidShortCode(code string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return regex.MatchString(code)
}
