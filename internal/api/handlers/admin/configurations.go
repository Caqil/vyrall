package admin

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// ConfigService interface for configuration operations
type ConfigService interface {
	GetAllConfigurations() (map[string]interface{}, error)
	GetConfigurationsByCategory(category string) (map[string]interface{}, error)
	UpdateConfiguration(key string, value interface{}) error
	GetConfigurationValue(key string) (interface{}, error)
	ResetConfigurationToDefault(key string) error
	ValidateConfiguration(key string, value interface{}) (bool, error)
	GetConfigurationHistory(key string, limit int) ([]map[string]interface{}, error)
}

// UpdateConfigRequest represents a request to update a configuration
type UpdateConfigRequest struct {
	Value interface{} `json:"value" binding:"required"`
}

// GetAllConfigurations returns all system configurations
func GetAllConfigurations(c *gin.Context) {
	configService := c.MustGet("configService").(ConfigService)

	configs, err := configService.GetAllConfigurations()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve configurations", err)
		return
	}

	response.Success(c, http.StatusOK, "Configurations retrieved successfully", configs)
}

// GetConfigurationsByCategory returns configurations for a specific category
func GetConfigurationsByCategory(c *gin.Context) {
	category := c.Param("category")
	configService := c.MustGet("configService").(ConfigService)

	configs, err := configService.GetConfigurationsByCategory(category)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve configurations", err)
		return
	}

	response.Success(c, http.StatusOK, "Configurations retrieved successfully", configs)
}

// GetConfigurationValue returns the value of a specific configuration
func GetConfigurationValue(c *gin.Context) {
	key := c.Param("key")
	configService := c.MustGet("configService").(ConfigService)

	value, err := configService.GetConfigurationValue(key)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve configuration value", err)
		return
	}

	response.Success(c, http.StatusOK, "Configuration retrieved successfully", gin.H{
		"key":   key,
		"value": value,
	})
}

// UpdateConfiguration updates a configuration value
func UpdateConfiguration(c *gin.Context) {
	key := c.Param("key")
	var req UpdateConfigRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	configService := c.MustGet("configService").(ConfigService)

	// Validate the configuration value
	valid, err := configService.ValidateConfiguration(key, req.Value)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to validate configuration", err)
		return
	}

	if !valid {
		response.Error(c, http.StatusBadRequest, "Invalid configuration value", nil)
		return
	}

	// Update the configuration
	if err := configService.UpdateConfiguration(key, req.Value); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update configuration", err)
		return
	}

	response.Success(c, http.StatusOK, "Configuration updated successfully", nil)
}

// ResetConfigurationToDefault resets a configuration to its default value
func ResetConfigurationToDefault(c *gin.Context) {
	key := c.Param("key")
	configService := c.MustGet("configService").(ConfigService)

	if err := configService.ResetConfigurationToDefault(key); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reset configuration", err)
		return
	}

	response.Success(c, http.StatusOK, "Configuration reset to default value successfully", nil)
}

// GetConfigurationHistory returns the history of changes for a configuration
func GetConfigurationHistory(c *gin.Context) {
	key := c.Param("key")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := parseInt(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	configService := c.MustGet("configService").(ConfigService)

	history, err := configService.GetConfigurationHistory(key, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve configuration history", err)
		return
	}

	response.Success(c, http.StatusOK, "Configuration history retrieved successfully", history)
}
