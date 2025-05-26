package business

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentService defines the interface for payment operations
type PaymentService interface {
	GetPaymentMethods(userID primitive.ObjectID) ([]models.PaymentMethod, error)
	AddPaymentMethod(userID primitive.ObjectID, method models.PaymentMethod) (primitive.ObjectID, error)
	DeletePaymentMethod(userID, methodID primitive.ObjectID) error
	SetDefaultPaymentMethod(userID, methodID primitive.ObjectID) error
	GetPaymentHistory(userID primitive.ObjectID, limit, offset int) ([]*models.Payment, int, error)
	GetPaymentDetails(userID, paymentID primitive.ObjectID) (*models.Payment, error)
	ProcessPayment(userID primitive.ObjectID, amount float64, currency string, description string, paymentMethodID primitive.ObjectID) (primitive.ObjectID, error)
	GetTransactionSummary(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetInvoice(userID, paymentID primitive.ObjectID) (string, error)
	GetSubscriptionPayments(userID, subscriptionID primitive.ObjectID, limit, offset int) ([]*models.Payment, int, error)
}

// GetPaymentMethods returns the user's saved payment methods
func GetPaymentMethods(c *gin.Context) {
	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Get payment methods
	methods, err := paymentService.GetPaymentMethods(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve payment methods", err)
		return
	}

	response.Success(c, http.StatusOK, "Payment methods retrieved successfully", methods)
}

// AddPaymentMethod adds a new payment method
func AddPaymentMethod(c *gin.Context) {
	var req struct {
		Type        string `json:"type" binding:"required"`
		Provider    string `json:"provider" binding:"required"`
		Token       string `json:"token" binding:"required"`
		Name        string `json:"name"`
		IsDefault   bool   `json:"is_default"`
		ExternalID  string `json:"external_id"`
		CardDetails struct {
			Last4       string `json:"last4"`
			Brand       string `json:"brand"`
			ExpiryMonth int    `json:"expiry_month"`
			ExpiryYear  int    `json:"expiry_year"`
			Country     string `json:"country"`
		} `json:"card_details"`
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

	// Validate payment method type
	validTypes := map[string]bool{
		"credit_card":   true,
		"bank_account":  true,
		"paypal":        true,
		"crypto_wallet": true,
	}

	if !validTypes[req.Type] {
		response.Error(c, http.StatusBadRequest, "Invalid payment method type", nil)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Create payment method object
	method := models.PaymentMethod{
		UserID:      userID.(primitive.ObjectID),
		Type:        req.Type,
		Provider:    req.Provider,
		ExternalID:  req.ExternalID,
		Name:        req.Name,
		Last4:       req.CardDetails.Last4,
		ExpiryMonth: req.CardDetails.ExpiryMonth,
		ExpiryYear:  req.CardDetails.ExpiryYear,
		Brand:       req.CardDetails.Brand,
		Country:     req.CardDetails.Country,
		IsDefault:   req.IsDefault,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add payment method
	methodID, err := paymentService.AddPaymentMethod(userID.(primitive.ObjectID), method)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add payment method", err)
		return
	}

	response.Success(c, http.StatusCreated, "Payment method added successfully", gin.H{
		"payment_method_id": methodID.Hex(),
	})
}

// DeletePaymentMethod removes a payment method
func DeletePaymentMethod(c *gin.Context) {
	methodID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment method ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Delete payment method
	if err := paymentService.DeletePaymentMethod(userID.(primitive.ObjectID), methodID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete payment method", err)
		return
	}

	response.Success(c, http.StatusOK, "Payment method deleted successfully", nil)
}

// SetDefaultPaymentMethod sets a payment method as default
func SetDefaultPaymentMethod(c *gin.Context) {
	methodID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment method ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Set default payment method
	if err := paymentService.SetDefaultPaymentMethod(userID.(primitive.ObjectID), methodID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to set default payment method", err)
		return
	}

	response.Success(c, http.StatusOK, "Default payment method set successfully", nil)
}

// GetPaymentHistory returns the user's payment history
func GetPaymentHistory(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Get payment history
	payments, total, err := paymentService.GetPaymentHistory(userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve payment history", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Payment history retrieved successfully", payments, limit, offset, total)
}

// GetPaymentDetails returns details for a specific payment
func GetPaymentDetails(c *gin.Context) {
	paymentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Get payment details
	payment, err := paymentService.GetPaymentDetails(userID.(primitive.ObjectID), paymentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve payment details", err)
		return
	}

	response.Success(c, http.StatusOK, "Payment details retrieved successfully", payment)
}

// ProcessPayment processes a new payment
func ProcessPayment(c *gin.Context) {
	var req struct {
		Amount          float64 `json:"amount" binding:"required"`
		Currency        string  `json:"currency" binding:"required"`
		Description     string  `json:"description" binding:"required"`
		PaymentMethodID string  `json:"payment_method_id" binding:"required"`
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

	// Parse payment method ID
	paymentMethodID, err := primitive.ObjectIDFromHex(req.PaymentMethodID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment method ID", err)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Process payment
	paymentID, err := paymentService.ProcessPayment(userID.(primitive.ObjectID), req.Amount, req.Currency, req.Description, paymentMethodID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process payment", err)
		return
	}

	response.Success(c, http.StatusOK, "Payment processed successfully", gin.H{
		"payment_id": paymentID.Hex(),
	})
}

// GetTransactionSummary returns a summary of transactions
func GetTransactionSummary(c *gin.Context) {
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

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Get transaction summary
	summary, err := paymentService.GetTransactionSummary(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve transaction summary", err)
		return
	}

	response.Success(c, http.StatusOK, "Transaction summary retrieved successfully", summary)
}

// GetInvoice returns an invoice for a payment
func GetInvoice(c *gin.Context) {
	paymentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Get invoice
	invoiceURL, err := paymentService.GetInvoice(userID.(primitive.ObjectID), paymentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve invoice", err)
		return
	}

	response.Success(c, http.StatusOK, "Invoice retrieved successfully", gin.H{
		"invoice_url": invoiceURL,
	})
}

// GetSubscriptionPayments returns payments for a subscription
func GetSubscriptionPayments(c *gin.Context) {
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

	paymentService := c.MustGet("paymentService").(PaymentService)

	// Get subscription payments
	payments, total, err := paymentService.GetSubscriptionPayments(userID.(primitive.ObjectID), subscriptionID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve subscription payments", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Subscription payments retrieved successfully", payments, limit, offset, total)
}
