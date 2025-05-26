package business

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CampaignService defines the interface for campaign operations
type CampaignService interface {
	CreateCampaign(campaign *models.Campaign) (primitive.ObjectID, error)
	GetCampaignByID(id primitive.ObjectID) (*models.Campaign, error)
	UpdateCampaign(campaign *models.Campaign) error
	DeleteCampaign(id primitive.ObjectID) error
	ListCampaigns(userID primitive.ObjectID, status string, limit, offset int) ([]*models.Campaign, int, error)
	GetCampaignPerformance(id primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetCampaignAds(id primitive.ObjectID, limit, offset int) ([]*models.Ad, int, error)
	ChangeCampaignStatus(id primitive.ObjectID, status string) error
}

// CreateCampaignRequest represents a request to create a campaign
type CreateCampaignRequest struct {
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description"`
	Objective     string            `json:"objective" binding:"required"`
	StartDate     time.Time         `json:"start_date" binding:"required"`
	EndDate       time.Time         `json:"end_date" binding:"required"`
	TotalBudget   float64           `json:"total_budget" binding:"required"`
	Currency      string            `json:"currency" binding:"required"`
	DailyBudget   float64           `json:"daily_budget"`
	Tags          []string          `json:"tags"`
	UTMParameters map[string]string `json:"utm_parameters"`
}

// CreateCampaign creates a new advertising campaign
func CreateCampaign(c *gin.Context) {
	var req CreateCampaignRequest
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

	// Validate objective
	validObjectives := map[string]bool{
		"awareness":     true,
		"consideration": true,
		"conversion":    true,
	}

	if !validObjectives[req.Objective] {
		response.Error(c, http.StatusBadRequest, "Invalid objective", nil)
		return
	}

	// Create campaign model
	campaign := &models.Campaign{
		AdvertiserID:    advertiserID.(primitive.ObjectID),
		Name:            req.Name,
		Description:     req.Description,
		Objective:       req.Objective,
		Status:          "draft",
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		TotalBudget:     req.TotalBudget,
		Currency:        req.Currency,
		SpentBudget:     0,
		RemainingBudget: req.TotalBudget,
		DailyBudget:     req.DailyBudget,
		AdCount:         0,
		Tags:            req.Tags,
		UTMParameters:   req.UTMParameters,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	campaignService := c.MustGet("campaignService").(CampaignService)

	// Create the campaign
	campaignID, err := campaignService.CreateCampaign(campaign)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create campaign", err)
		return
	}

	response.Success(c, http.StatusCreated, "Campaign created successfully", gin.H{
		"campaign_id": campaignID.Hex(),
	})
}

// GetCampaign retrieves campaign details
func GetCampaign(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid campaign ID", err)
		return
	}

	campaignService := c.MustGet("campaignService").(CampaignService)

	campaign, err := campaignService.GetCampaignByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign", err)
		return
	}

	// Check if the authenticated user is the campaign owner
	userID, _ := c.Get("userID")
	if campaign.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to access this campaign", nil)
		return
	}

	response.Success(c, http.StatusOK, "Campaign retrieved successfully", campaign)
}

// UpdateCampaign updates an existing campaign
func UpdateCampaign(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid campaign ID", err)
		return
	}

	var req CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	campaignService := c.MustGet("campaignService").(CampaignService)

	// Get the existing campaign
	campaign, err := campaignService.GetCampaignByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign", err)
		return
	}

	// Check if the authenticated user is the campaign owner
	userID, _ := c.Get("userID")
	if campaign.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to update this campaign", nil)
		return
	}

	// Check if campaign can be updated
	if campaign.Status != "draft" && campaign.Status != "paused" {
		response.Error(c, http.StatusBadRequest, "Only draft or paused campaigns can be updated", nil)
		return
	}

	// Update campaign fields
	campaign.Name = req.Name
	campaign.Description = req.Description
	campaign.Objective = req.Objective
	campaign.StartDate = req.StartDate
	campaign.EndDate = req.EndDate
	campaign.TotalBudget = req.TotalBudget
	campaign.Currency = req.Currency
	campaign.RemainingBudget = req.TotalBudget - campaign.SpentBudget
	campaign.DailyBudget = req.DailyBudget
	campaign.Tags = req.Tags
	campaign.UTMParameters = req.UTMParameters
	campaign.UpdatedAt = time.Now()

	// Update the campaign
	if err := campaignService.UpdateCampaign(campaign); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update campaign", err)
		return
	}

	response.Success(c, http.StatusOK, "Campaign updated successfully", gin.H{
		"campaign_id": id.Hex(),
	})
}

// DeleteCampaign deletes a campaign
func DeleteCampaign(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid campaign ID", err)
		return
	}

	campaignService := c.MustGet("campaignService").(CampaignService)

	// Get the existing campaign
	campaign, err := campaignService.GetCampaignByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign", err)
		return
	}

	// Check if the authenticated user is the campaign owner
	userID, _ := c.Get("userID")
	if campaign.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to delete this campaign", nil)
		return
	}

	// Delete the campaign
	if err := campaignService.DeleteCampaign(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete campaign", err)
		return
	}

	response.Success(c, http.StatusOK, "Campaign deleted successfully", nil)
}

// ListCampaigns retrieves campaigns for the authenticated user
func ListCampaigns(c *gin.Context) {
	status := c.DefaultQuery("status", "all")
	limit, offset := getPaginationParams(c)

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	campaignService := c.MustGet("campaignService").(CampaignService)

	campaigns, total, err := campaignService.ListCampaigns(userID.(primitive.ObjectID), status, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaigns", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Campaigns retrieved successfully", campaigns, limit, offset, total)
}

// GetCampaignPerformance retrieves campaign performance metrics
func GetCampaignPerformance(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid campaign ID", err)
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

	campaignService := c.MustGet("campaignService").(CampaignService)

	// Get the existing campaign
	campaign, err := campaignService.GetCampaignByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign", err)
		return
	}

	// Check if the authenticated user is the campaign owner
	userID, _ := c.Get("userID")
	if campaign.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to access this campaign's performance", nil)
		return
	}

	// Get campaign performance
	performance, err := campaignService.GetCampaignPerformance(id, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign performance", err)
		return
	}

	response.Success(c, http.StatusOK, "Campaign performance retrieved successfully", performance)
}

// GetCampaignAds retrieves ads for a campaign
func GetCampaignAds(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid campaign ID", err)
		return
	}

	limit, offset := getPaginationParams(c)

	campaignService := c.MustGet("campaignService").(CampaignService)

	// Get the existing campaign
	campaign, err := campaignService.GetCampaignByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign", err)
		return
	}

	// Check if the authenticated user is the campaign owner
	userID, _ := c.Get("userID")
	if campaign.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to access this campaign's ads", nil)
		return
	}

	// Get campaign ads
	ads, total, err := campaignService.GetCampaignAds(id, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign ads", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Campaign ads retrieved successfully", ads, limit, offset, total)
}

// ChangeCampaignStatus changes the status of a campaign
func ChangeCampaignStatus(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid campaign ID", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	campaignService := c.MustGet("campaignService").(CampaignService)

	// Get the existing campaign
	campaign, err := campaignService.GetCampaignByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve campaign", err)
		return
	}

	// Check if the authenticated user is the campaign owner
	userID, _ := c.Get("userID")
	if campaign.AdvertiserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to change this campaign's status", nil)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft":     true,
		"active":    true,
		"paused":    true,
		"completed": true,
	}

	if !validStatuses[req.Status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	// Change campaign status
	if err := campaignService.ChangeCampaignStatus(id, req.Status); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to change campaign status", err)
		return
	}

	response.Success(c, http.StatusOK, "Campaign status changed successfully", nil)
}
