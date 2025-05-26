package validation

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// StringLength checks if a string's length is between min and max (inclusive)
func StringLength(s string, min, max int) bool {
	length := utf8.RuneCountInString(s)
	return length >= min && length <= max
}

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// AlphanumericWithSymbols checks if a string contains only alphanumeric characters and allowed symbols
func AlphanumericWithSymbols(s string, allowedSymbols string) bool {
	if IsEmpty(s) {
		return false
	}

	// Build regex pattern
	pattern := "^[a-zA-Z0-9" + regexp.QuoteMeta(allowedSymbols) + "]+$"
	match, _ := regexp.MatchString(pattern, s)
	return match
}

// AlphanumericOnly checks if a string contains only alphanumeric characters
func AlphanumericOnly(s string) bool {
	return AlphanumericWithSymbols(s, "")
}

// ContainsAny checks if a string contains any of the provided characters
func ContainsAny(s string, chars string) bool {
	return strings.ContainsAny(s, chars)
}

// ContainsAll checks if a string contains all of the provided characters
func ContainsAll(s string, chars string) bool {
	for _, char := range chars {
		if !strings.ContainsRune(s, char) {
			return false
		}
	}
	return true
}

// NumberInRange checks if a number is within a specified range (inclusive)
func NumberInRange(num, min, max int) bool {
	return num >= min && num <= max
}

// MatchesPattern checks if a string matches a regular expression pattern
func MatchesPattern(s string, pattern string) bool {
	match, _ := regexp.MatchString(pattern, s)
	return match
}

// IsURL checks if a string is a valid URL
func IsURL(s string) bool {
	// Simple URL regex pattern
	pattern := `^(https?:\/\/)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\/[a-zA-Z0-9_\-\.~!\*'\(\);:@&=\+\$,\/\?%#\[\]]*)?$`
	return MatchesPattern(s, pattern)
}

// IsDate checks if a string is a valid date in YYYY-MM-DD format
func IsDate(s string) bool {
	pattern := `^\d{4}-\d{2}-\d{2}$`
	if !MatchesPattern(s, pattern) {
		return false
	}

	// Additional date validation could be performed here
	// For now, we're just checking the format
	return true
}

// IsBooleanString checks if a string represents a boolean value
func IsBooleanString(s string) bool {
	lowerS := strings.ToLower(s)
	return lowerS == "true" || lowerS == "false" || lowerS == "1" || lowerS == "0" || lowerS == "yes" || lowerS == "no"
}

// SanitizeString removes or replaces potentially unsafe characters
func SanitizeString(s string) string {
	// Replace HTML tags with spaces
	htmlTagPattern := "<[^>]*>"
	re := regexp.MustCompile(htmlTagPattern)
	s = re.ReplaceAllString(s, " ")

	// Remove control characters
	controlCharPattern := "[\x00-\x1F\x7F]"
	re = regexp.MustCompile(controlCharPattern)
	s = re.ReplaceAllString(s, "")

	return strings.TrimSpace(s)
}
