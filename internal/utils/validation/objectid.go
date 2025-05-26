package validation

import (
	"encoding/hex"
	"regexp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IsValidObjectID checks if a string is a valid MongoDB ObjectID
func IsValidObjectID(id string) bool {
	// Check if the string is empty
	if IsEmpty(id) {
		return false
	}

	// Check length
	if len(id) != 24 {
		return false
	}

	// Check if it contains only hexadecimal characters
	pattern := "^[0-9a-fA-F]{24}$"
	return MatchesPattern(id, pattern)
}

// ToObjectID converts a string to a MongoDB ObjectID
func ToObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

// IsNilObjectID checks if an ObjectID is nil/zero
func IsNilObjectID(id primitive.ObjectID) bool {
	return id == primitive.NilObjectID
}

// ValidateObjectIDs checks if a list of string IDs are all valid ObjectIDs
func ValidateObjectIDs(ids []string) bool {
	for _, id := range ids {
		if !IsValidObjectID(id) {
			return false
		}
	}
	return true
}

// ConvertToObjectIDs converts a list of string IDs to ObjectIDs
func ConvertToObjectIDs(ids []string) ([]primitive.ObjectID, error) {
	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objectIDs[i] = objectID
	}
	return objectIDs, nil
}

// GenerateObjectID generates a new MongoDB ObjectID
func GenerateObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// ObjectIDToHex converts an ObjectID to a hexadecimal string
func ObjectIDToHex(id primitive.ObjectID) string {
	return id.Hex()
}

// IsObjectIDInList checks if an ObjectID is present in a list of ObjectIDs
func IsObjectIDInList(id primitive.ObjectID, list []primitive.ObjectID) bool {
	for _, item := range list {
		if item == id {
			return true
		}
	}
	return false
}

// IsValidObjectIDPattern checks if a string follows the pattern of an ObjectID
// but doesn't validate if it's a properly generated one
func IsValidObjectIDPattern(id string) bool {
	// Check for common formats like:
	// - 24 hex characters
	// - 12 byte representation
	// - "ObjectId(...)" format

	if IsEmpty(id) {
		return false
	}

	// Check for standard 24-char hex format
	if len(id) == 24 {
		if _, err := hex.DecodeString(id); err == nil {
			return true
		}
	}

	// Check for ObjectId(...) format
	objectIdPattern := `^ObjectId\(['"](0-9a-fA-F){24}['"]\)$`
	if MatchesPattern(id, objectIdPattern) {
		return true
	}

	return false
}

// ExtractObjectIDFromString extracts an ObjectID from a string that might contain other text
func ExtractObjectIDFromString(s string) string {
	// Extract using regex
	pattern := `[0-9a-fA-F]{24}`
	re := regexp.MustCompile(pattern)
	match := re.FindString(s)

	if match != "" && IsValidObjectID(match) {
		return match
	}

	// Check for ObjectId(...) format
	objectIdPattern := `ObjectId\(['"]([0-9a-fA-F]{24})['"]`
	re = regexp.MustCompile(objectIdPattern)
	matches := re.FindStringSubmatch(s)

	if len(matches) > 1 && IsValidObjectID(matches[1]) {
		return matches[1]
	}

	return ""
}
