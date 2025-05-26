package validation

import (
	"crypto/subtle"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordOptions defines options for password validation
type PasswordOptions struct {
	MinLength         int
	MaxLength         int
	RequireUppercase  bool
	RequireLowercase  bool
	RequireNumbers    bool
	RequireSymbols    bool
	DisallowedStrings []string
}

// DefaultPasswordOptions returns default password validation options
func DefaultPasswordOptions() PasswordOptions {
	return PasswordOptions{
		MinLength:         8,
		MaxLength:         100,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireNumbers:    true,
		RequireSymbols:    true,
		DisallowedStrings: []string{"password", "123456", "qwerty"},
	}
}

// ValidatePassword validates a password against the specified options
func ValidatePassword(password string, options PasswordOptions) (bool, string) {
	// Check length
	if len(password) < options.MinLength {
		return false, fmt.Sprintf("Password must be at least %d characters long", options.MinLength)
	}

	if len(password) > options.MaxLength {
		return false, fmt.Sprintf("Password must not exceed %d characters", options.MaxLength)
	}

	// Check for disallowed strings
	lowerPassword := strings.ToLower(password)
	for _, disallowed := range options.DisallowedStrings {
		if strings.Contains(lowerPassword, strings.ToLower(disallowed)) {
			return false, fmt.Sprintf("Password contains disallowed string: %s", disallowed)
		}
	}

	// Check for uppercase letters if required
	if options.RequireUppercase {
		hasUpper := false
		for _, c := range password {
			if unicode.IsUpper(c) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return false, "Password must contain at least one uppercase letter"
		}
	}

	// Check for lowercase letters if required
	if options.RequireLowercase {
		hasLower := false
		for _, c := range password {
			if unicode.IsLower(c) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return false, "Password must contain at least one lowercase letter"
		}
	}

	// Check for numbers if required
	if options.RequireNumbers {
		hasNumber := false
		for _, c := range password {
			if unicode.IsDigit(c) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return false, "Password must contain at least one number"
		}
	}

	// Check for symbols if required
	if options.RequireSymbols {
		hasSymbol := false
		for _, c := range password {
			if !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsSpace(c) {
				hasSymbol = true
				break
			}
		}
		if !hasSymbol {
			return false, "Password must contain at least one symbol"
		}
	}

	return true, ""
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePasswordAndHash compares a password with a hash
func ComparePasswordAndHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// PasswordStrength assesses the strength of a password from 0 (weak) to 100 (strong)
func PasswordStrength(password string) int {
	// Base score
	score := 0

	// Length checks
	length := len(password)
	if length < 8 {
		score += length * 2 // Up to 14 points
	} else if length < 16 {
		score += 16 + (length - 8) // Up to 23 points
	} else {
		score += 24 // Maximum 24 points for length
	}

	// Character type checks
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSymbol := false

	for _, c := range password {
		if unicode.IsLower(c) {
			hasLower = true
		} else if unicode.IsUpper(c) {
			hasUpper = true
		} else if unicode.IsDigit(c) {
			hasDigit = true
		} else if !unicode.IsSpace(c) {
			hasSymbol = true
		}
	}

	// Add points for different character types
	charTypeCount := 0
	if hasLower {
		charTypeCount++
	}
	if hasUpper {
		charTypeCount++
	}
	if hasDigit {
		charTypeCount++
	}
	if hasSymbol {
		charTypeCount++
	}

	score += charTypeCount * 10 // Up to 40 points

	// Check for common patterns
	lowerPassword := strings.ToLower(password)

	// Subtract points for common words
	commonWords := []string{"password", "123456", "qwerty", "admin", "welcome", "letmein"}
	for _, word := range commonWords {
		if strings.Contains(lowerPassword, word) {
			score -= 20
			break
		}
	}

	// Ensure score is within 0-100 range
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

// ConstantTimeCompare compares strings in constant time to prevent timing attacks
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// GenerateRandomPassword generates a random password with the specified options
func GenerateRandomPassword(options PasswordOptions) string {
	// This is a simplified implementation
	// In a real application, use a more secure random generator

	const (
		lowerChars = "abcdefghijklmnopqrstuvwxyz"
		upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numChars   = "0123456789"
		symChars   = "!@#$%^&*()-_=+[]{}|;:,.<>?/"
	)

	// Calculate password length between min and max
	length := options.MinLength
	if options.MaxLength > options.MinLength {
		length = options.MinLength + (options.MaxLength-options.MinLength)/2
	}

	// Ensure length is at least 8
	if length < 8 {
		length = 8
	}

	// Generate password
	password := ""
	allChars := ""

	// Include required character types
	if options.RequireLowercase {
		password += string(lowerChars[len(password)%len(lowerChars)])
		allChars += lowerChars
	}
	if options.RequireUppercase {
		password += string(upperChars[len(password)%len(upperChars)])
		allChars += upperChars
	}
	if options.RequireNumbers {
		password += string(numChars[len(password)%len(numChars)])
		allChars += numChars
	}
	if options.RequireSymbols {
		password += string(symChars[len(password)%len(symChars)])
		allChars += symChars
	}

	// Fill the rest with random characters
	for len(password) < length {
		password += string(allChars[len(password)%len(allChars)])
	}

	return password
}
