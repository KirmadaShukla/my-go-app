package tutor

import (
	"fmt"
	"strconv"
	"strings"
)

// Subjects is the single source of truth for allowed learning areas.
var Subjects = []string{"maths", "science", "english", "activities"}

func IsValidSubject(subject string) bool {
	s := strings.ToLower(strings.TrimSpace(subject))
	for _, allowed := range Subjects {
		if s == allowed {
			return true
		}
	}
	return false
}

// NormalizeClass maps values like "3", "Class 3", "3rd" to "3" for classes 1-10.
func NormalizeClass(raw string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "class", "")
	s = strings.ReplaceAll(s, "std", "")
	s = strings.ReplaceAll(s, "standard", "")
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "stndrh")
	s = strings.TrimSpace(s)

	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 10 {
		return "", fmt.Errorf("child class must be between 1 and 10")
	}
	return strconv.Itoa(n), nil
}
