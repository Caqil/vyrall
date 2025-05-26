package admin

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BackupService interface for backup operations
type BackupService interface {
	CreateBackup(name, description string, includeMedia bool) (primitive.ObjectID, error)
	GetBackups(limit, offset int) ([]map[string]interface{}, int, error)
	GetBackupByID(id primitive.ObjectID) (map[string]interface{}, error)
	DeleteBackup(id primitive.ObjectID) error
	RestoreFromBackup(id primitive.ObjectID) error
	GetBackupStatus(id primitive.ObjectID) (string, error)
	ScheduleBackup(name, description string, includeMedia bool, schedule string) (primitive.ObjectID, error)
	UpdateBackupSchedule(id primitive.ObjectID, schedule string) error
	DeleteBackupSchedule(id primitive.ObjectID) error
}

// BackupRequest represents a request to create a backup
type BackupRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	IncludeMedia bool   `json:"include_media"`
}

// BackupScheduleRequest represents a request to schedule a backup
type BackupScheduleRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	IncludeMedia bool   `json:"include_media"`
	Schedule     string `json:"schedule" binding:"required"` // Cron expression
}

// CreateBackup creates a new database backup
func CreateBackup(c *gin.Context) {
	var req BackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	backupID, err := backupService.CreateBackup(req.Name, req.Description, req.IncludeMedia)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create backup", err)
		return
	}

	response.Success(c, http.StatusCreated, "Backup created successfully", gin.H{
		"backup_id": backupID.Hex(),
	})
}

// GetBackups returns a list of available backups
func GetBackups(c *gin.Context) {
	limit, offset := getPaginationParams(c)
	backupService := c.MustGet("backupService").(BackupService)

	backups, total, err := backupService.GetBackups(limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve backups", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Backups retrieved successfully", backups, limit, offset, total)
}

// GetBackupDetail returns details about a specific backup
func GetBackupDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid backup ID", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	backup, err := backupService.GetBackupByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve backup details", err)
		return
	}

	response.Success(c, http.StatusOK, "Backup details retrieved successfully", backup)
}

// DeleteBackup deletes a backup
func DeleteBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid backup ID", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	if err := backupService.DeleteBackup(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete backup", err)
		return
	}

	response.Success(c, http.StatusOK, "Backup deleted successfully", nil)
}

// RestoreBackup restores the system from a backup
func RestoreBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid backup ID", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	if err := backupService.RestoreFromBackup(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to restore from backup", err)
		return
	}

	response.Success(c, http.StatusOK, "System restored from backup successfully", nil)
}

// GetBackupStatus returns the status of a backup operation
func GetBackupStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid backup ID", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	status, err := backupService.GetBackupStatus(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve backup status", err)
		return
	}

	response.Success(c, http.StatusOK, "Backup status retrieved successfully", gin.H{
		"status": status,
	})
}

// ScheduleBackup schedules a recurring backup
func ScheduleBackup(c *gin.Context) {
	var req BackupScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	scheduleID, err := backupService.ScheduleBackup(req.Name, req.Description, req.IncludeMedia, req.Schedule)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to schedule backup", err)
		return
	}

	response.Success(c, http.StatusCreated, "Backup scheduled successfully", gin.H{
		"schedule_id": scheduleID.Hex(),
	})
}

// UpdateBackupSchedule updates a backup schedule
func UpdateBackupSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid schedule ID", err)
		return
	}

	var req struct {
		Schedule string `json:"schedule" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	if err := backupService.UpdateBackupSchedule(id, req.Schedule); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update backup schedule", err)
		return
	}

	response.Success(c, http.StatusOK, "Backup schedule updated successfully", nil)
}

// DeleteBackupSchedule deletes a backup schedule
func DeleteBackupSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid schedule ID", err)
		return
	}

	backupService := c.MustGet("backupService").(BackupService)

	if err := backupService.DeleteBackupSchedule(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete backup schedule", err)
		return
	}

	response.Success(c, http.StatusOK, "Backup schedule deleted successfully", nil)
}

// Helper function to get pagination parameters
func getPaginationParams(c *gin.Context) (int, int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := parseInt(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := parseInt(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}
