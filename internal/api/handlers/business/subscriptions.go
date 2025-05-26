package business

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionService defines the interface for subscription operations
type SubscriptionService interface {
	GetSubscriptionPlans() ([]models.SubscriptionPlan, error)
	GetSubscriptionPlanByID(id primitive.ObjectID) (*models.SubscriptionPlan, error)
	GetActiveSubscription(userID primitive.ObjectID) (*models.Subscription, error)
	CreateSubscription(userID, planID, paymentMethodID primitive.ObjectID) (primitive.ObjectID, error)
	CancelSubscription(userID, subscriptionID primitive.ObjectID, cancelAtPeriodEnd bool) error
	UpdateSubscription(userID, subscriptionID, newPlanID primitive.ObjectID) error
	PauseSubscription(userID, subscriptionID primitive.ObjectID, resumeAt time.Time) error
	ResumeSubscription(userID, subscriptionID primitive.ObjectID) error
	GetSubscriptionHistory(userID primitive.ObjectID, limit, offset int) ([]*models.Subscription, int, error)
	GetSubscriptionDetails(userID, subscriptionID primitive.ObjectID) (*models.Subscription, error)
	GetSubscriptionInvoices(userID, subscriptionID primitive.ObjectID, limit, offset int) ([]map[string]interface{}, int, error)
	ManageNotifications(userID, subscriptionID primitive.ObjectID, settings map[string]bool) error
}

// GetSubscriptionPlans returns available subscription plans
func GetSubscriptionPlans(c *gin.Context) {
	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Get subscription plans
	plans, err := subscriptionService.GetSubscriptionPlans()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve subscription plans", err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription plans retrieved successfully", plans)
}

// GetSubscriptionPlan returns details for a specific subscription plan
func GetSubscriptionPlan(c *gin.Context) {
	planID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid plan ID", err)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Get subscription plan
	plan, err := subscriptionService.GetSubscriptionPlanByID(planID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve subscription plan", err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription plan retrieved successfully", plan)
}

// GetActiveSubscription returns the user's active subscription
func GetActiveSubscription(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Get active subscription
	subscription, err := subscriptionService.GetActiveSubscription(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve active subscription", err)
		return
	}

	// Check if the user has an active subscription
	if subscription == nil {
		response.Success(c, http.StatusOK, "No active subscription found", nil)
		return
	}

	response.Success(c, http.StatusOK, "Active subscription retrieved successfully", subscription)
}

// CreateSubscription subscribes to a plan
func CreateSubscription(c *gin.Context) {
	var req struct {
		PlanID          string `json:"plan_id" binding:"required"`
		PaymentMethodID string `json:"payment_method_id" binding:"required"`
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

	// Parse IDs
	planID, err := primitive.ObjectIDFromHex(req.PlanID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid plan ID", err)
		return
	}

	paymentMethodID, err := primitive.ObjectIDFromHex(req.PaymentMethodID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment method ID", err)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Check if the user already has an active subscription
	activeSubscription, err := subscriptionService.GetActiveSubscription(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check active subscription", err)
		return
	}

	if activeSubscription != nil {
		response.Error(c, http.StatusBadRequest, "User already has an active subscription", nil)
		return
	}

	// Create subscription
	subscriptionID, err := subscriptionService.CreateSubscription(userID.(primitive.ObjectID), planID, paymentMethodID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create subscription", err)
		return
	}

	response.Success(c, http.StatusCreated, "Subscription created successfully", gin.H{
		"subscription_id": subscriptionID.Hex(),
	})
}

// CancelSubscription cancels a subscription
func CancelSubscription(c *gin.Context) {
	subscriptionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid subscription ID", err)
		return
	}

	var req struct {
		CancelAtPeriodEnd bool `json:"cancel_at_period_end"`
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

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Cancel subscription
	if err := subscriptionService.CancelSubscription(userID.(primitive.ObjectID), subscriptionID, req.CancelAtPeriodEnd); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to cancel subscription", err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription canceled successfully", nil)
}

// UpdateSubscription changes the subscription plan
func UpdateSubscription(c *gin.Context) {
	subscriptionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid subscription ID", err)
		return
	}

	var req struct {
		NewPlanID string `json:"new_plan_id" binding:"required"`
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

	// Parse new plan ID
	newPlanID, err := primitive.ObjectIDFromHex(req.NewPlanID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid plan ID", err)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Update subscription
	if err := subscriptionService.UpdateSubscription(userID.(primitive.ObjectID), subscriptionID, newPlanID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update subscription", err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription updated successfully", nil)
}

// PauseSubscription pauses a subscription
func PauseSubscription(c *gin.Context) {
	subscriptionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid subscription ID", err)
		return
	}

	var req struct {
		ResumeAt string `json:"resume_at" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Parse resume date
	resumeAt, err := time.Parse("2006-01-02", req.ResumeAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid date format for resume_at", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Pause subscription
	if err := subscriptionService.PauseSubscription(userID.(primitive.ObjectID), subscriptionID, resumeAt); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to pause subscription", err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription paused successfully", nil)
}

// ResumeSubscription resumes a paused subscription
func ResumeSubscription(c *gin.Context) {
	subscriptionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid subscription ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Resume subscription
	if err := subscriptionService.ResumeSubscription(userID.(primitive.ObjectID), subscriptionID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to resume subscription", err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription resumed successfully", nil)
}

// GetSubscriptionHistory returns the user's subscription history
func GetSubscriptionHistory(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Get subscription history
	subscriptions, total, err := subscriptionService.GetSubscriptionHistory(userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve subscription history", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Subscription history retrieved successfully", subscriptions, limit, offset, total)
}

// GetSubscriptionDetails returns details for a specific subscription
func GetSubscriptionDetails(c *gin.Context) {
	subscriptionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid subscription ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Get subscription details
	subscription, err := subscriptionService.GetSubscriptionDetails(userID.(primitive.ObjectID), subscriptionID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve subscription details", err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription details retrieved successfully", subscription)
}

// GetSubscriptionInvoices returns invoices for a subscription
func GetSubscriptionInvoices(c *gin.Context) {
	subscriptionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid subscription ID", err)
		return
	}

	limit, offset := getPaginationParams(c)

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Get subscription invoices
	invoices, total, err := subscriptionService.GetSubscriptionInvoices(userID.(primitive.ObjectID), subscriptionID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve subscription invoices", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Subscription invoices retrieved successfully", invoices, limit, offset, total)
}

// ManageNotifications updates notification settings for a subscription
func ManageNotifications(c *gin.Context) {
	subscriptionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid subscription ID", err)
		return
	}

	var req struct {
		Settings map[string]bool `json:"settings" binding:"required"`
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

	subscriptionService := c.MustGet("subscriptionService").(SubscriptionService)

	// Validate settings
	requiredSettings := []string{"renewal_reminder", "payment_failed", "price_change"}
	for _, setting := range requiredSettings {
		if _, ok := req.Settings[setting]; !ok {
			response.Error(c, http.StatusBadRequest, "Missing required setting: "+setting, nil)
			return
		}
	}

	// Update notification settings
	if err := subscriptionService.ManageNotifications(userID.(primitive.ObjectID), subscriptionID, req.Settings); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notification settings", err)
		return
	}

	response.Success(c, http.StatusOK, "Notification settings updated successfully", nil)
}
