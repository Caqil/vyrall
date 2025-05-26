package validation

import (
	"net"
	"strings"
)

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) bool {
	// Check if email is empty
	if IsEmpty(email) {
		return false
	}

	// Check for length
	if !StringLength(email, 3, 254) {
		return false
	}

	// Pattern for basic email validation
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	if !MatchesPattern(email, pattern) {
		return false
	}

	// Split email into parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	// Check local part length
	if !StringLength(parts[0], 1, 64) {
		return false
	}

	// Check domain part length
	if !StringLength(parts[1], 1, 255) {
		return false
	}

	// Check domain has at least one dot
	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}

// ValidateEmailDomain checks if the domain part of an email has valid MX records
func ValidateEmailDomain(email string) bool {
	if !ValidateEmail(email) {
		return false
	}

	// Extract domain from email
	parts := strings.Split(email, "@")
	domain := parts[1]

	// Check for MX records
	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		return false
	}

	return true
}

// IsDisposableEmail checks if an email is from a disposable email provider
func IsDisposableEmail(email string) bool {
	if !ValidateEmail(email) {
		return false
	}

	// Extract domain from email
	parts := strings.Split(email, "@")
	domain := parts[1]

	// List of common disposable email domains
	disposableDomains := []string{
		"10minutemail.com",
		"disposablemail.com",
		"mailinator.com",
		"guerrillamail.com",
		"temp-mail.org",
		"tempmail.com",
		"throwawaymail.com",
		"fakeinbox.com",
		"getnada.com",
		"yopmail.com",
		// Add more as needed
	}

	for _, disposableDomain := range disposableDomains {
		if domain == disposableDomain || strings.HasSuffix(domain, "."+disposableDomain) {
			return true
		}
	}

	return false
}

// NormalizeEmail converts an email address to lowercase and removes unnecessary parts
func NormalizeEmail(email string) string {
	if !ValidateEmail(email) {
		return email
	}

	// Convert to lowercase
	email = strings.ToLower(email)

	// Extract local and domain parts
	parts := strings.Split(email, "@")
	local := parts[0]
	domain := parts[1]

	// Remove dots from Gmail addresses (since Gmail ignores dots)
	if domain == "gmail.com" || domain == "googlemail.com" {
		// Remove dots from local part
		local = strings.ReplaceAll(local, ".", "")

		// Remove everything after + in Gmail addresses
		if idx := strings.Index(local, "+"); idx != -1 {
			local = local[:idx]
		}

		// Normalize Gmail domain
		domain = "gmail.com"
	}

	return local + "@" + domain
}

// IsBusinessEmail checks if an email is from a business domain (not a free email provider)
func IsBusinessEmail(email string) bool {
	if !ValidateEmail(email) {
		return false
	}

	// Extract domain from email
	parts := strings.Split(email, "@")
	domain := parts[1]

	// List of common free email providers
	freeEmailProviders := []string{
		"gmail.com",
		"yahoo.com",
		"hotmail.com",
		"outlook.com",
		"aol.com",
		"icloud.com",
		"protonmail.com",
		"mail.com",
		"zoho.com",
		"gmx.com",
		// Add more as needed
	}

	for _, provider := range freeEmailProviders {
		if domain == provider || strings.HasSuffix(domain, "."+provider) {
			return false
		}
	}

	return true
}
