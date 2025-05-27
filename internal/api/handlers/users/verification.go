package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VerificationHandler handles user verification operations
type VerificationHandler struct {
	userService *user.Service
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(userService *user.Service) *VerificationHandler {
	return &VerificationHandler{
		userService: userService,
	}
}

// VerifyEmail handles the request to verify a user's email
func (h *VerificationHandler) VerifyEmail(c *gin.Context) {
	// Get verification code from URL parameter
	code := c.Query("code")
	if code == "" {
		response.ValidationError(c, "Verification code is required", nil)
		return
	}

	// Verify the email
	err := h.userService.VerifyEmail(c.Request.Context(), code)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify email", err)
		return
	}

	// Return success response
	response.OK(c, "Email verified successfully", nil)
}

// SendVerificationEmail handles the request to send a verification email
func (h *VerificationHandler) SendVerificationEmail(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Send verification email
	err := h.userService.SendVerificationEmail(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send verification email", err)
		return
	}

	// Return success response
	response.OK(c, "Verification email sent", nil)
}

// VerifyPhone handles the request to verify a user's phone number
func (h *VerificationHandler) VerifyPhone(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if req.Code == "" {
		response.ValidationError(c, "Verification code is required", nil)
		return
	}

	// Verify the phone number
	err := h.userService.VerifyPhone(c.Request.Context(), userID.(primitive.ObjectID), req.Code)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify phone number", err)
		return
	}

	// Return success response
	response.OK(c, "Phone number verified successfully", nil)
}

// SendPhoneVerification handles the request to send a phone verification code
func (h *VerificationHandler) SendPhoneVerification(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if req.PhoneNumber == "" {
		response.ValidationError(c, "Phone number is required", nil)
		return
	}

	// Send phone verification
	err := h.userService.SendPhoneVerification(c.Request.Context(), userID.(primitive.ObjectID), req.PhoneNumber)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send phone verification", err)
		return
	}

	// Return success response
	response.OK(c, "Verification code sent to phone", nil)
}

// RequestVerification handles the request to verify account (blue badge)
func (h *VerificationHandler) RequestVerification(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Category       string   `json:"category" binding:"required"` // celebrity, brand, etc.
		FullName       string   `json:"full_name" binding:"required"`
		DocumentType   string   `json:"document_type" binding:"required"` // passport, license, etc.
		DocumentImages []string `json:"document_images" binding:"required"`
		AdditionalInfo string   `json:"additional_info,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.DocumentImages) == 0 {
		response.ValidationError(c, "At least one document image is required", nil)
		return
	}

	// Submit verification request
	requestID, err := h.userService.RequestVerification(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		req.Category,
		req.FullName,
		req.DocumentType,
		req.DocumentImages,
		req.AdditionalInfo,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to submit verification request", err)
		return
	}

	// Return success response
	response.Created(c, "Verification request submitted successfully", gin.H{
		"request_id": requestID.Hex(),
		"status":     "pending",
	})
}

// GetVerificationStatus handles the request to get the status of a verification request
func (h *VerificationHandler) GetVerificationStatus(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get verification status
	status, err := h.userService.GetVerificationStatus(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get verification status", err)
		return
	}

	// Return success response
	response.OK(c, "Verification status retrieved successfully", status)
}
