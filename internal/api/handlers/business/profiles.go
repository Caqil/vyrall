package business

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BusinessProfileService defines the interface for business profile operations
type BusinessProfileService interface {
	GetBusinessProfile(userID primitive.ObjectID) (map[string]interface{}, error)
	CreateBusinessProfile(userID primitive.ObjectID, profile map[string]interface{}) error
	UpdateBusinessProfile(userID primitive.ObjectID, profile map[string]interface{}) error
	GetBusinessCategories() ([]string, error)
	VerifyBusinessProfile(userID primitive.ObjectID, documents []string) (string, error)
	GetVerificationStatus(userID primitive.ObjectID) (map[string]interface{}, error)
	GetBusinessHours(userID primitive.ObjectID) (map[string]interface{}, error)
	UpdateBusinessHours(userID primitive.ObjectID, hours map[string]interface{}) error
	GetBusinessLocations(userID primitive.ObjectID) ([]map[string]interface{}, error)
	AddBusinessLocation(userID primitive.ObjectID, location map[string]interface{}) (primitive.ObjectID, error)
	UpdateBusinessLocation(userID, locationID primitive.ObjectID, location map[string]interface{}) error
	DeleteBusinessLocation(userID, locationID primitive.ObjectID) error
	GetBusinessContact(userID primitive.ObjectID) (map[string]interface{}, error)
	UpdateBusinessContact(userID primitive.ObjectID, contact map[string]interface{}) error
}

// GetBusinessProfile returns the business profile for the authenticated user
func GetBusinessProfile(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Get business profile
	profile, err := businessProfileService.GetBusinessProfile(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve business profile", err)
		return
	}

	response.Success(c, http.StatusOK, "Business profile retrieved successfully", profile)
}

// CreateBusinessProfile creates a new business profile
func CreateBusinessProfile(c *gin.Context) {
	var profile map[string]interface{}
	if err := c.ShouldBindJSON(&profile); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Validate required fields
	requiredFields := []string{"name", "category", "description", "website"}
	for _, field := range requiredFields {
		if _, ok := profile[field]; !ok {
			response.Error(c, http.StatusBadRequest, "Missing required field: "+field, nil)
			return
		}
	}

	// Create business profile
	if err := businessProfileService.CreateBusinessProfile(userID.(primitive.ObjectID), profile); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create business profile", err)
		return
	}

	response.Success(c, http.StatusCreated, "Business profile created successfully", nil)
}

// UpdateBusinessProfile updates an existing business profile
func UpdateBusinessProfile(c *gin.Context) {
	var profile map[string]interface{}
	if err := c.ShouldBindJSON(&profile); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Update business profile
	if err := businessProfileService.UpdateBusinessProfile(userID.(primitive.ObjectID), profile); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update business profile", err)
		return
	}

	response.Success(c, http.StatusOK, "Business profile updated successfully", nil)
}

// GetBusinessCategories returns available business categories
func GetBusinessCategories(c *gin.Context) {
	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Get business categories
	categories, err := businessProfileService.GetBusinessCategories()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve business categories", err)
		return
	}

	response.Success(c, http.StatusOK, "Business categories retrieved successfully", categories)
}

// VerifyBusinessProfile submits a business profile for verification
func VerifyBusinessProfile(c *gin.Context) {
	var req struct {
		Documents []string `json:"documents" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Submit profile for verification
	verificationID, err := businessProfileService.VerifyBusinessProfile(userID.(primitive.ObjectID), req.Documents)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to submit verification request", err)
		return
	}

	response.Success(c, http.StatusOK, "Verification request submitted successfully", gin.H{
		"verification_id": verificationID,
	})
}

// GetVerificationStatus returns the verification status of a business profile
func GetVerificationStatus(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Get verification status
	status, err := businessProfileService.GetVerificationStatus(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve verification status", err)
		return
	}

	response.Success(c, http.StatusOK, "Verification status retrieved successfully", status)
}

// GetBusinessHours returns the business hours
func GetBusinessHours(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Get business hours
	hours, err := businessProfileService.GetBusinessHours(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve business hours", err)
		return
	}

	response.Success(c, http.StatusOK, "Business hours retrieved successfully", hours)
}

// UpdateBusinessHours updates the business hours
func UpdateBusinessHours(c *gin.Context) {
	var hours map[string]interface{}
	if err := c.ShouldBindJSON(&hours); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Update business hours
	if err := businessProfileService.UpdateBusinessHours(userID.(primitive.ObjectID), hours); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update business hours", err)
		return
	}

	response.Success(c, http.StatusOK, "Business hours updated successfully", nil)
}

// GetBusinessLocations returns the business locations
func GetBusinessLocations(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Get business locations
	locations, err := businessProfileService.GetBusinessLocations(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve business locations", err)
		return
	}

	response.Success(c, http.StatusOK, "Business locations retrieved successfully", locations)
}

// AddBusinessLocation adds a new business location
func AddBusinessLocation(c *gin.Context) {
	var location map[string]interface{}
	if err := c.ShouldBindJSON(&location); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Validate required fields
	requiredFields := []string{"name", "address", "city", "country"}
	for _, field := range requiredFields {
		if _, ok := location[field]; !ok {
			response.Error(c, http.StatusBadRequest, "Missing required field: "+field, nil)
			return
		}
	}

	// Add business location
	locationID, err := businessProfileService.AddBusinessLocation(userID.(primitive.ObjectID), location)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add business location", err)
		return
	}

	response.Success(c, http.StatusCreated, "Business location added successfully", gin.H{
		"location_id": locationID.Hex(),
	})
}

// UpdateBusinessLocation updates a business location
func UpdateBusinessLocation(c *gin.Context) {
	locationID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid location ID", err)
		return
	}

	var location map[string]interface{}
	if err := c.ShouldBindJSON(&location); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Update business location
	if err := businessProfileService.UpdateBusinessLocation(userID.(primitive.ObjectID), locationID, location); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update business location", err)
		return
	}

	response.Success(c, http.StatusOK, "Business location updated successfully", nil)
}

// DeleteBusinessLocation deletes a business location
func DeleteBusinessLocation(c *gin.Context) {
	locationID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid location ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Delete business location
	if err := businessProfileService.DeleteBusinessLocation(userID.(primitive.ObjectID), locationID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete business location", err)
		return
	}

	response.Success(c, http.StatusOK, "Business location deleted successfully", nil)
}

// GetBusinessContact returns the business contact information
func GetBusinessContact(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Get business contact
	contact, err := businessProfileService.GetBusinessContact(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve business contact", err)
		return
	}

	response.Success(c, http.StatusOK, "Business contact retrieved successfully", contact)
}

// UpdateBusinessContact updates the business contact information
func UpdateBusinessContact(c *gin.Context) {
	var contact map[string]interface{}
	if err := c.ShouldBindJSON(&contact); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	businessProfileService := c.MustGet("businessProfileService").(BusinessProfileService)

	// Update business contact
	if err := businessProfileService.UpdateBusinessContact(userID.(primitive.ObjectID), contact); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update business contact", err)
		return
	}

	response.Success(c, http.StatusOK, "Business contact updated successfully", nil)
}
