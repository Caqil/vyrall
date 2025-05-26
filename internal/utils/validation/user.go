package validation

import (
	"strings"
	"time"
	"unicode"
)

// UsernameOptions defines options for username validation
type UsernameOptions struct {
	MinLength         int
	MaxLength         int
	AllowedCharacters string
	ReservedUsernames []string
}

// DefaultUsernameOptions returns default username validation options
func DefaultUsernameOptions() UsernameOptions {
	return UsernameOptions{
		MinLength:         3,
		MaxLength:         30,
		AllowedCharacters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.",
		ReservedUsernames: []string{
			"admin", "administrator", "system", "moderator", "mod", "support",
			"help", "root", "staff", "official", "security", "info", "contact",
			"about", "terms", "privacy", "login", "register", "signup", "signin",
			"signout", "logout", "password", "username", "account", "profile",
			"settings", "notifications", "messages", "explore", "search", "trending",
			"popular", "home", "timeline", "feed", "discover", "news", "events",
		},
	}
}

// ValidateUsername validates a username against the specified options
func ValidateUsername(username string, options UsernameOptions) (bool, string) {
	// Check if username is empty
	if IsEmpty(username) {
		return false, "Username cannot be empty"
	}

	// Check length
	if len(username) < options.MinLength {
		return false, "Username must be at least " + string(rune('0'+options.MinLength)) + " characters long"
	}

	if len(username) > options.MaxLength {
		return false, "Username cannot exceed " + string(rune('0'+options.MaxLength)) + " characters"
	}

	// Check for invalid characters
	for _, c := range username {
		if !strings.ContainsRune(options.AllowedCharacters, c) {
			return false, "Username contains invalid character: " + string(c)
		}
	}

	// Check for reserved usernames
	lowerUsername := strings.ToLower(username)
	for _, reserved := range options.ReservedUsernames {
		if lowerUsername == strings.ToLower(reserved) {
			return false, "This username is reserved and cannot be used"
		}
	}

	// Username cannot start or end with period or underscore
	if strings.HasPrefix(username, ".") || strings.HasPrefix(username, "_") ||
		strings.HasSuffix(username, ".") || strings.HasSuffix(username, "_") {
		return false, "Username cannot start or end with period or underscore"
	}

	// Username cannot contain consecutive periods or underscores
	if strings.Contains(username, "..") || strings.Contains(username, "__") {
		return false, "Username cannot contain consecutive periods or underscores"
	}

	return true, ""
}

// ValidateDisplayName validates a user's display name
func ValidateDisplayName(displayName string) (bool, string) {
	// Check if display name is empty
	if IsEmpty(displayName) {
		return false, "Display name cannot be empty"
	}

	// Check length
	nameLen := len(strings.TrimSpace(displayName))
	if nameLen < 1 {
		return false, "Display name is too short"
	}

	if nameLen > 50 {
		return false, "Display name cannot exceed 50 characters"
	}

	// Check for invalid characters
	for _, c := range displayName {
		if c < 32 || c == 127 { // Control characters
			return false, "Display name contains invalid character"
		}
	}

	return true, ""
}

// ValidateBio validates a user's bio
func ValidateBio(bio string) (bool, string) {
	// Bio can be empty, but not longer than the maximum length
	if len(bio) > 500 {
		return false, "Bio cannot exceed 500 characters"
	}

	return true, ""
}

// ValidateAge checks if a user's age is within acceptable range
func ValidateAge(birthdate time.Time) (bool, string) {
	now := time.Now()

	// Check for future date
	if birthdate.After(now) {
		return false, "Birth date cannot be in the future"
	}

	// Calculate age
	age := now.Year() - birthdate.Year()

	// Adjust age if birthday hasn't occurred yet this year
	if now.Month() < birthdate.Month() || (now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		age--
	}

	// Check minimum age (13 is common for social platforms)
	if age < 13 {
		return false, "User must be at least 13 years old"
	}

	// Check for unreasonable age (150 is unlikely to be a real person)
	if age > 150 {
		return false, "Birth date is too far in the past"
	}

	return true, ""
}

// IsValidLocation checks if a location string is valid
func IsValidLocation(location string) bool {
	// Location can be empty
	if IsEmpty(location) {
		return true
	}

	// Check length
	if len(location) > 100 {
		return false
	}

	// Check for invalid characters
	for _, c := range location {
		if c < 32 || c == 127 { // Control characters
			return false
		}
	}

	return true
}

// ValidateWebsite checks if a website URL is valid
func ValidateWebsite(website string) (bool, string) {
	// Website can be empty
	if IsEmpty(website) {
		return true, ""
	}

	// Check length
	if len(website) > 200 {
		return false, "Website URL is too long"
	}

	// Check if it's a valid URL
	if !IsURL(website) {
		return false, "Website is not a valid URL"
	}

	return true, ""
}

// ValidatePhoneNumber checks if a phone number is valid
func ValidatePhoneNumber(phone string) (bool, string) {
	// Phone can be empty
	if IsEmpty(phone) {
		return true, ""
	}

	// Remove common formatting characters
	cleanPhone := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '+' {
			return r
		}
		return -1
	}, phone)

	// Check length
	if len(cleanPhone) < 7 || len(cleanPhone) > 15 {
		return false, "Phone number has invalid length"
	}

	// Basic format check
	if !strings.HasPrefix(cleanPhone, "+") && len(cleanPhone) < 10 {
		return false, "International phone numbers should start with + and have at least 10 digits"
	}

	return true, ""
}

// ValidateGender checks if a gender value is valid
func ValidateGender(gender string) bool {
	// Gender can be empty
	if IsEmpty(gender) {
		return true
	}

	// Common gender options
	validGenders := map[string]bool{
		"male":              true,
		"female":            true,
		"non-binary":        true,
		"other":             true,
		"prefer not to say": true,
	}

	return validGenders[strings.ToLower(gender)]
}
