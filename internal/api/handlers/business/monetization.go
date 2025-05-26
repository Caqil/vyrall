package business

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MonetizationService defines the interface for monetization operations
type MonetizationService interface {
	GetMonetizationStatus(userID primitive.ObjectID) (map[string]interface{}, error)
	ApplyForMonetization(userID primitive.ObjectID, application map[string]interface{}) error
	GetMonetizationOptions(userID primitive.ObjectID) ([]map[string]interface{}, error)
	EnableMonetizationFeature(userID primitive.ObjectID, feature string, settings map[string]interface{}) error
	DisableMonetizationFeature(userID primitive.ObjectID, feature string) error
	GetEarnings(userID primitive.ObjectID, period string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetEarningsByFeature(userID primitive.ObjectID, feature string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetPayoutHistory(userID primitive.ObjectID, limit, offset int) ([]map[string]interface{}, int, error)
	RequestPayout(userID primitive.ObjectID, amount float64, method string) (string, error)
	GetTaxInformation(userID primitive.ObjectID) (map[string]interface{}, error)
	UpdateTaxInformation(userID primitive.ObjectID, taxInfo map[string]interface{}) error
}

// GetMonetizationStatus returns the current monetization status for the user
func GetMonetizationStatus(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Get monetization status
	status, err := monetizationService.GetMonetizationStatus(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve monetization status", err)
		return
	}

	response.Success(c, http.StatusOK, "Monetization status retrieved successfully", status)
}

// ApplyForMonetization submits an application for monetization
func ApplyForMonetization(c *gin.Context) {
	var application map[string]interface{}
	if err := c.ShouldBindJSON(&application); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Apply for monetization
	if err := monetizationService.ApplyForMonetization(userID.(primitive.ObjectID), application); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to submit monetization application", err)
		return
	}

	response.Success(c, http.StatusOK, "Monetization application submitted successfully", nil)
}

// GetMonetizationOptions returns available monetization options
func GetMonetizationOptions(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Get monetization options
	options, err := monetizationService.GetMonetizationOptions(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve monetization options", err)
		return
	}

	response.Success(c, http.StatusOK, "Monetization options retrieved successfully", options)
}

// EnableMonetizationFeature enables a monetization feature
func EnableMonetizationFeature(c *gin.Context) {
	feature := c.Param("feature")

	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Validate feature
	validFeatures := map[string]bool{
		"ads":           true,
		"subscriptions": true,
		"tips":          true,
		"merchandise":   true,
		"sponsorships":  true,
	}

	if !validFeatures[feature] {
		response.Error(c, http.StatusBadRequest, "Invalid monetization feature", nil)
		return
	}

	// Enable monetization feature
	if err := monetizationService.EnableMonetizationFeature(userID.(primitive.ObjectID), feature, settings); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to enable monetization feature", err)
		return
	}

	response.Success(c, http.StatusOK, "Monetization feature enabled successfully", nil)
}

// DisableMonetizationFeature disables a monetization feature
func DisableMonetizationFeature(c *gin.Context) {
	feature := c.Param("feature")

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Validate feature
	validFeatures := map[string]bool{
		"ads":           true,
		"subscriptions": true,
		"tips":          true,
		"merchandise":   true,
		"sponsorships":  true,
	}

	if !validFeatures[feature] {
		response.Error(c, http.StatusBadRequest, "Invalid monetization feature", nil)
		return
	}

	// Disable monetization feature
	if err := monetizationService.DisableMonetizationFeature(userID.(primitive.ObjectID), feature); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to disable monetization feature", err)
		return
	}

	response.Success(c, http.StatusOK, "Monetization feature disabled successfully", nil)
}

// GetEarnings returns earnings information
func GetEarnings(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly")

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Validate period
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"yearly":  true,
		"custom":  true,
	}

	if !validPeriods[period] {
		response.Error(c, http.StatusBadRequest, "Invalid period", nil)
		return
	}

	// Get earnings
	earnings, err := monetizationService.GetEarnings(userID.(primitive.ObjectID), period, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve earnings", err)
		return
	}

	response.Success(c, http.StatusOK, "Earnings retrieved successfully", earnings)
}

// GetEarningsByFeature returns earnings information for a specific feature
func GetEarningsByFeature(c *gin.Context) {
	feature := c.Param("feature")

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Validate feature
	validFeatures := map[string]bool{
		"ads":           true,
		"subscriptions": true,
		"tips":          true,
		"merchandise":   true,
		"sponsorships":  true,
	}

	if !validFeatures[feature] {
		response.Error(c, http.StatusBadRequest, "Invalid monetization feature", nil)
		return
	}

	// Get earnings by feature
	earnings, err := monetizationService.GetEarningsByFeature(userID.(primitive.ObjectID), feature, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve earnings for feature", err)
		return
	}

	response.Success(c, http.StatusOK, "Feature earnings retrieved successfully", earnings)
}

// GetPayoutHistory returns payout history
func GetPayoutHistory(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Get payout history
	payouts, total, err := monetizationService.GetPayoutHistory(userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve payout history", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Payout history retrieved successfully", payouts, limit, offset, total)
}

// RequestPayout requests a payout
func RequestPayout(c *gin.Context) {
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		Method string  `json:"method" binding:"required"`
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

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Validate method
	validMethods := map[string]bool{
		"bank_transfer": true,
		"paypal":        true,
		"stripe":        true,
		"crypto":        true,
	}

	if !validMethods[req.Method] {
		response.Error(c, http.StatusBadRequest, "Invalid payout method", nil)
		return
	}

	// Request payout
	payoutID, err := monetizationService.RequestPayout(userID.(primitive.ObjectID), req.Amount, req.Method)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to request payout", err)
		return
	}

	response.Success(c, http.StatusOK, "Payout requested successfully", gin.H{
		"payout_id": payoutID,
	})
}

// GetTaxInformation returns tax information
func GetTaxInformation(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Get tax information
	taxInfo, err := monetizationService.GetTaxInformation(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve tax information", err)
		return
	}

	response.Success(c, http.StatusOK, "Tax information retrieved successfully", taxInfo)
}

// UpdateTaxInformation updates tax information
func UpdateTaxInformation(c *gin.Context) {
	var taxInfo map[string]interface{}
	if err := c.ShouldBindJSON(&taxInfo); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	monetizationService := c.MustGet("monetizationService").(MonetizationService)

	// Update tax information
	if err := monetizationService.UpdateTaxInformation(userID.(primitive.ObjectID), taxInfo); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update tax information", err)
		return
	}

	response.Success(c, http.StatusOK, "Tax information updated successfully", nil)
}
