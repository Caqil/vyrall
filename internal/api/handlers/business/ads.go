package business

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdService defines the interface for ad operations
type AdService interface {
	CreateAd(ad *models.Ad) (primitive.ObjectID, error)
	GetAdByID(id primitive.ObjectID) (*models.Ad, error)
	UpdateAd(ad *models.Ad) error
	DeleteAd(id primitive.ObjectID) error
	ListAds(userID primitive.ObjectID, status string, limit, offset int) ([]*models.Ad, int, error)
	SubmitAdForReview(id primitive.ObjectID) error
	GetAdPerformance(id primitive.ObjectID, startDate, endDate time.Time) (*models.AdMetrics, error)
	EstimateAdAudience(targeting models.AdTargeting) (map[string]interface{}, error)
	GetAdPlacements() ([]map[string]interface{}, error)
}

// CreateAdRequest represents a request to create an ad
type CreateAdRequest struct {
	Title          string              `json:"title" binding:"required"`
	Description    string              `json:"description" binding:"required"`
	MediaFiles     []string            `json:"media_files"`
	DestinationURL string              `json:"destination_url" binding:"required,url"`
	CallToAction   string              `json:"call_to_action"`
	Format         string              `json:"format" binding:"required"`
	Placement      string              `json:"placement" binding:"required"`
	TargetAudience models.AdTargeting  `json:"target_audience"`
	Budget         models.AdBudget     `json:"budget" binding:"required"`
	StartDate      time.Time           `json:"start_date" binding:"required"`
	EndDate        time.Time           `json:"end_date" binding:"required"`
	Schedule       []models.AdSchedule `json:"schedule"`
	Keywords       []string            `json:"keywords"`
	CampaignID     string              `json:"campaign_id"`
}

// CreateAd creates a new ad
func CreateAd(c *gin.Context) {
	var req CreateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get advertiser ID from authenticated user
	advertiserID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Parse campaign ID if provided
	var campaignID primitive.ObjectID
	var err error
	if req.CampaignID != "" {
		campaignID, err = primitive.ObjectIDFromHex(req.CampaignID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid campaign ID", err)
			return
		}
	}

	// Validate dates
	now := time.Now()
	if req.StartDate.Before(now) {
		response.Error(c, http.StatusBadRequest, "Start date must be in the future", nil)
		return
	}

	if req.EndDate.Before(req.StartDate) {
		response.Error(c, http.StatusBadRequest, "End date must be after start date", nil)
		return
	}

	// Convert media file strings to Media objects
	var mediaFiles []models.Media
	for _, mediaID := range req.MediaFiles {
		mediaObjectID, err := primitive.ObjectIDFromHex(mediaID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid media ID: "+mediaID, err)
			return
		}
		mediaFiles = append(mediaFiles, models.Media{ID: mediaObjectID})
	}

	// Create ad model
	ad := &models.Ad{
		AdvertiserID:   advertiserID.(primitive.ObjectID),
		CampaignID:     campaignID,
		Title:          req.Title,
		Description:    req.Description,
		MediaFiles:     mediaFiles,
		DestinationURL: req.DestinationURL,
		CallToAction:   req.CallToAction,
		Status:         "draft",
		Format:         req.Format,
		Placement:      req.Placement,
		TargetAudience: req.TargetAudience,
		Budget:         req.Budget,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		Schedule:       req.Schedule,
		Keywords:       req.Keywords,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	adService := c.MustGet("adService").(AdService)

	// Create the ad
	adID, err := adService.CreateAd(ad)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create ad", err)
		return
	}

	response.Success(c, http.StatusCreated, "Ad created successfully", gin.H{
		"ad_id": adID.Hex(),
	})
}

// GetAd retrieves ad details
func GetAd(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ad ID", err)
		return
	}

	adService := c.MustGet("adService").(AdService)

	ad, err := adService.GetAdByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad", err)
		return
	}

	// Check if the authenticated user is the ad owner
	userID, _ := c.Get("userID")
	if ad.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to access this ad", nil)
		return
	}

	response.Success(c, http.StatusOK, "Ad retrieved successfully", ad)
}

// UpdateAd updates an existing ad
func UpdateAd(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ad ID", err)
		return
	}

	var req CreateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	adService := c.MustGet("adService").(AdService)

	// Get the existing ad
	ad, err := adService.GetAdByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad", err)
		return
	}

	// Check if the authenticated user is the ad owner
	userID, _ := c.Get("userID")
	if ad.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to update this ad", nil)
		return
	}

	// Check if ad can be updated
	if ad.Status != "draft" && ad.Status != "rejected" {
		response.Error(c, http.StatusBadRequest, "Only draft or rejected ads can be updated", nil)
		return
	}

	// Convert media file strings to Media objects
	var mediaFiles []models.Media
	for _, mediaID := range req.MediaFiles {
		mediaObjectID, err := primitive.ObjectIDFromHex(mediaID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid media ID: "+mediaID, err)
			return
		}
		mediaFiles = append(mediaFiles, models.Media{ID: mediaObjectID})
	}

	// Update ad fields
	ad.Title = req.Title
	ad.Description = req.Description
	ad.MediaFiles = mediaFiles
	ad.DestinationURL = req.DestinationURL
	ad.CallToAction = req.CallToAction
	ad.Format = req.Format
	ad.Placement = req.Placement
	ad.TargetAudience = req.TargetAudience
	ad.Budget = req.Budget
	ad.StartDate = req.StartDate
	ad.EndDate = req.EndDate
	ad.Schedule = req.Schedule
	ad.Keywords = req.Keywords
	ad.UpdatedAt = time.Now()

	// Update the ad
	if err := adService.UpdateAd(ad); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update ad", err)
		return
	}

	response.Success(c, http.StatusOK, "Ad updated successfully", gin.H{
		"ad_id": id.Hex(),
	})
}

// DeleteAd deletes an ad
func DeleteAd(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ad ID", err)
		return
	}

	adService := c.MustGet("adService").(AdService)

	// Get the existing ad
	ad, err := adService.GetAdByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad", err)
		return
	}

	// Check if the authenticated user is the ad owner
	userID, _ := c.Get("userID")
	if ad.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to delete this ad", nil)
		return
	}

	// Delete the ad
	if err := adService.DeleteAd(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete ad", err)
		return
	}

	response.Success(c, http.StatusOK, "Ad deleted successfully", nil)
}

// ListAds retrieves ads for the authenticated user
func ListAds(c *gin.Context) {
	status := c.DefaultQuery("status", "all")
	limit, offset := getPaginationParams(c)

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	adService := c.MustGet("adService").(AdService)

	ads, total, err := adService.ListAds(userID.(primitive.ObjectID), status, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ads", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Ads retrieved successfully", ads, limit, offset, total)
}

// SubmitAdForReview submits an ad for review
func SubmitAdForReview(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ad ID", err)
		return
	}

	adService := c.MustGet("adService").(AdService)

	// Get the existing ad
	ad, err := adService.GetAdByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad", err)
		return
	}

	// Check if the authenticated user is the ad owner
	userID, _ := c.Get("userID")
	if ad.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to submit this ad for review", nil)
		return
	}

	// Check if ad can be submitted
	if ad.Status != "draft" && ad.Status != "rejected" {
		response.Error(c, http.StatusBadRequest, "Only draft or rejected ads can be submitted for review", nil)
		return
	}

	// Submit the ad for review
	if err := adService.SubmitAdForReview(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to submit ad for review", err)
		return
	}

	response.Success(c, http.StatusOK, "Ad submitted for review successfully", nil)
}

// GetAdPerformance retrieves ad performance metrics
func GetAdPerformance(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ad ID", err)
		return
	}

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

	adService := c.MustGet("adService").(AdService)

	// Get the existing ad
	ad, err := adService.GetAdByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad", err)
		return
	}

	// Check if the authenticated user is the ad owner
	userID, _ := c.Get("userID")
	if ad.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to access this ad's performance", nil)
		return
	}

	// Get ad performance
	metrics, err := adService.GetAdPerformance(id, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad performance", err)
		return
	}

	response.Success(c, http.StatusOK, "Ad performance retrieved successfully", metrics)
}

// EstimateAdAudience estimates the potential audience for ad targeting
func EstimateAdAudience(c *gin.Context) {
	var targeting models.AdTargeting
	if err := c.ShouldBindJSON(&targeting); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	adService := c.MustGet("adService").(AdService)

	// Get audience estimate
	estimate, err := adService.EstimateAdAudience(targeting)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to estimate ad audience", err)
		return
	}

	response.Success(c, http.StatusOK, "Ad audience estimated successfully", estimate)
}

// GetAdPlacements retrieves available ad placements
func GetAdPlacements(c *gin.Context) {
	adService := c.MustGet("adService").(AdService)

	// Get ad placements
	placements, err := adService.GetAdPlacements()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad placements", err)
		return
	}

	response.Success(c, http.StatusOK, "Ad placements retrieved successfully", placements)
}

// Helper function to get pagination parameters
func getPaginationParams(c *gin.Context) (int, int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit := 10
	if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	offset := 0
	if parsedOffset, err := parseInt(offsetStr); err == nil && parsedOffset >= 0 {
		offset = parsedOffset
	}

	return limit, offset
}

// Helper function to parse int from string
func parseInt(str string) (int, error) {
	var value int
	_, err := fmt.Sscanf(str, "%d", &value)
	return value, err
}
