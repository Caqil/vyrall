package response

import (
	"time"
)

// MetadataResponse represents the common metadata for API responses
type MetadataResponse struct {
	APIVersion     string          `json:"api_version"`
	ServerTime     time.Time       `json:"server_time"`
	ExecutionTime  float64         `json:"execution_time_ms,omitempty"`
	RequestID      string          `json:"request_id,omitempty"`
	PaginationInfo *PaginationInfo `json:"pagination,omitempty"`
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"` // Unix timestamp
}

// CreateMetadata creates a new metadata object
func CreateMetadata(apiVersion string, executionTime float64, requestID string) MetadataResponse {
	return MetadataResponse{
		APIVersion:    apiVersion,
		ServerTime:    time.Now(),
		ExecutionTime: executionTime,
		RequestID:     requestID,
	}
}

// CreateMetadataWithPagination creates metadata that includes pagination information
func CreateMetadataWithPagination(apiVersion string, executionTime float64, requestID string, pagination *PaginationInfo) MetadataResponse {
	metadata := CreateMetadata(apiVersion, executionTime, requestID)
	metadata.PaginationInfo = pagination
	return metadata
}

// AddRateLimit adds rate limit information to the metadata
func AddRateLimit(metadata *MetadataResponse, limit, remaining int, reset int64) {
	// Convert to map if not already
	metadataMap, ok := metadata.(*map[string]interface{})
	if !ok {
		m := make(map[string]interface{})
		m["api_version"] = metadata.APIVersion
		m["server_time"] = metadata.ServerTime
		m["execution_time_ms"] = metadata.ExecutionTime
		m["request_id"] = metadata.RequestID
		if metadata.PaginationInfo != nil {
			m["pagination"] = metadata.PaginationInfo
		}
		metadataMap = &m
	}

	// Add rate limit info
	(*metadataMap)["rate_limit"] = RateLimitInfo{
		Limit:     limit,
		Remaining: remaining,
		Reset:     reset,
	}
}

// AddCustomMetadata adds a custom field to the metadata
func AddCustomMetadata(metadata *MetadataResponse, key string, value interface{}) {
	// Convert to map if not already
	metadataMap, ok := metadata.(*map[string]interface{})
	if !ok {
		m := make(map[string]interface{})
		m["api_version"] = metadata.APIVersion
		m["server_time"] = metadata.ServerTime
		m["execution_time_ms"] = metadata.ExecutionTime
		m["request_id"] = metadata.RequestID
		if metadata.PaginationInfo != nil {
			m["pagination"] = metadata.PaginationInfo
		}
		metadataMap = &m
	}

	// Add custom field
	(*metadataMap)[key] = value
}
