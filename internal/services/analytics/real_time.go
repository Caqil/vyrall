package analytics

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/Caqil/vyrall/internal/database"
	"github.com/Caqil/vyrall/internal/utils/logger"
	"go.mongodb.org/mongo-driver/bson"
)

// RealTimeService provides real-time analytics functionality
type RealTimeService struct {
	db           *database.Database
	cache        *database.RedisClient
	log          *logger.Logger
	counters     map[string]int
	counterMutex sync.RWMutex
	// Channel for publishing real-time events
	eventChan chan *RealTimeEvent
}

// RealTimeEvent represents a real-time analytics event
type RealTimeEvent struct {
	Type       string                 `json:"type"`
	EntityID   string                 `json:"entity_id,omitempty"`
	EntityType string                 `json:"entity_type,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// NewRealTimeService creates a new real-time analytics service
func NewRealTimeService(db *database.Database, cache *database.RedisClient, log *logger.Logger) *RealTimeService {
	service := &RealTimeService{
		db:        db,
		cache:     cache,
		log:       log,
		counters:  make(map[string]int),
		eventChan: make(chan *RealTimeEvent, 1000), // Buffer size of 1000
	}

	// Start the event processor
	go service.processEvents()

	// Start the counter flusher
	go service.flushCounters()

	return service
}

// ProcessEvent processes an analytics event in real-time
func (s *RealTimeService) ProcessEvent(ctx context.Context, event *Event) error {
	// Increment the appropriate counters based on the event type
	s.incrementCounters(event)

	// Create real-time event
	rtEvent := &RealTimeEvent{
		Type:       event.EventType,
		EntityID:   event.EntityID,
		EntityType: event.EntityType,
		UserID:     event.UserID,
		Timestamp:  event.Timestamp,
		Data:       event.Properties,
	}

	// Send to the event channel (non-blocking)
	select {
	case s.eventChan <- rtEvent:
		// Event sent successfully
	default:
		// Channel is full, log the issue but don't block
		s.log.Warn("Real-time event channel is full, dropping event", "event_type", event.EventType)
	}

	return nil
}

// GetActiveUsers gets the number of active users in real-time
func (s *RealTimeService) GetActiveUsers(ctx context.Context) (int, error) {
	// Try to get from cache
	cacheKey := "active_users:count"
	cachedCount, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedCount != "" {
		count, err := strconv.Atoi(cachedCount)
		if err == nil {
			return count, nil
		}
	}

	// Calculate active users (users active in the last 15 minutes)
	fifteenMinutesAgo := time.Now().Add(-15 * time.Minute)

	// Query for active sessions
	filter := bson.M{
		"last_activity": bson.M{
			"$gte": fifteenMinutesAgo,
		},
	}

	count, err := s.db.CountDocuments(ctx, "user_sessions", filter)
	if err != nil {
		s.log.Error("Failed to count active users", "error", err)
		return 0, err
	}

	// Cache the result for 1 minute
	s.cache.SetWithExpiration(ctx, cacheKey, strconv.Itoa(int(count)), 1*time.Minute)

	return int(count), nil
}

// GetRealtimeMetrics gets real-time metrics
func (s *RealTimeService) GetRealtimeMetrics(ctx context.Context) (map[string]interface{}, error) {
	// Get metrics from both cache and current counters
	metrics := make(map[string]interface{})

	// Get active users
	activeUsers, err := s.GetActiveUsers(ctx)
	if err != nil {
		s.log.Error("Failed to get active users", "error", err)
		// Continue despite error
	}
	metrics["active_users"] = activeUsers

	// Get current counter values
	s.counterMutex.RLock()
	counters := make(map[string]int)
	for key, value := range s.counters {
		counters[key] = value
	}
	s.counterMutex.RUnlock()

	// Add counter values to metrics
	metrics["counters"] = counters

	// Get cached metrics
	cachedMetrics, err := s.getCachedMetrics(ctx)
	if err != nil {
		s.log.Error("Failed to get cached metrics", "error", err)
		// Continue despite error
	} else {
		// Merge cached metrics with current metrics
		for key, value := range cachedMetrics {
			metrics[key] = value
		}
	}

	return metrics, nil
}

// SubscribeToEvents allows clients to subscribe to real-time events
func (s *RealTimeService) SubscribeToEvents(ctx context.Context, entityType, entityID string) (<-chan *RealTimeEvent, error) {
	// Create a channel for the subscriber
	eventChan := make(chan *RealTimeEvent, 100)

	// Create a subscription key
	subscriptionKey := entityType
	if entityID != "" {
		subscriptionKey += ":" + entityID
	}

	// Store the subscription in Redis
	// In a real implementation, this would be more complex to handle multiple subscribers
	err := s.cache.SetWithExpiration(ctx, "subscription:"+subscriptionKey, "active", 1*time.Hour)
	if err != nil {
		s.log.Error("Failed to store subscription", "error", err, "key", subscriptionKey)
		return nil, err
	}

	// Start a goroutine to filter and forward events to the subscriber
	go func() {
		defer close(eventChan)

		// Create a new channel to receive events from the main event channel
		sub := make(chan *RealTimeEvent, 100)

		// Register the subscriber
		// In a real implementation, this would use a proper pub/sub mechanism

		// Listen for events until the context is canceled
		for {
			select {
			case <-ctx.Done():
				// Context canceled, clean up and return
				return
			case event := <-sub:
				// Check if the event matches the subscription
				if entityType == "" || event.EntityType == entityType {
					if entityID == "" || event.EntityID == entityID {
						// Send the event to the subscriber
						select {
						case eventChan <- event:
							// Event sent successfully
						default:
							// Subscriber's channel is full, log and continue
							s.log.Warn("Subscriber channel is full, dropping event", "event_type", event.Type)
						}
					}
				}
			}
		}
	}()

	return eventChan, nil
}

// incrementCounters increments the appropriate counters based on the event
func (s *RealTimeService) incrementCounters(event *Event) {
	s.counterMutex.Lock()
	defer s.counterMutex.Unlock()

	// Increment total events counter
	s.counters["total_events"]++

	// Increment counter for this event type
	s.counters["event_type:"+event.EventType]++

	// Increment entity-specific counters if applicable
	if event.EntityID != "" && event.EntityType != "" {
		s.counters["entity:"+event.EntityType+":"+event.EntityID]++
	}

	// Increment user-specific counters if applicable
	if event.UserID != "" {
		s.counters["user:"+event.UserID]++
	}
}

// flushCounters periodically flushes counters to the database
func (s *RealTimeService) flushCounters() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.doFlushCounters()
	}
}

// doFlushCounters performs the actual counter flushing
func (s *RealTimeService) doFlushCounters() {
	// Get the current counters
	s.counterMutex.Lock()
	counters := make(map[string]int)
	for key, value := range s.counters {
		counters[key] = value
		s.counters[key] = 0 // Reset counter
	}
	s.counterMutex.Unlock()

	// Store in cache for quick access
	ctx := context.Background()
	for key, value := range counters {
		cacheKey := "realtime:counter:" + key
		currentValue, err := s.cache.Get(ctx, cacheKey)
		var newValue int
		if err == nil && currentValue != "" {
			if existingValue, err := strconv.Atoi(currentValue); err == nil {
				newValue = existingValue + value
			} else {
				newValue = value
			}
		} else {
			newValue = value
		}

		// Store updated value with 1-hour expiration
		s.cache.SetWithExpiration(ctx, cacheKey, strconv.Itoa(newValue), 1*time.Hour)
	}

	// Log flush event
	s.log.Debug("Flushed real-time counters", "counter_count", len(counters))
}

// processEvents processes events from the event channel
func (s *RealTimeService) processEvents() {
	for event := range s.eventChan {
		// Process the event (e.g., publish to subscribers)
		s.publishEvent(event)
	}
}

// publishEvent publishes an event to subscribers
func (s *RealTimeService) publishEvent(event *RealTimeEvent) {
	// In a real implementation, this would use a proper pub/sub mechanism
	// For now, we'll just log the event
	s.log.Debug("Real-time event", "type", event.Type, "entity_type", event.EntityType, "entity_id", event.EntityID)
}

// getCachedMetrics gets metrics from the cache
func (s *RealTimeService) getCachedMetrics(ctx context.Context) (map[string]interface{}, error) {
	// Get all counter keys
	pattern := "realtime:counter:*"
	keys, err := s.cache.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]interface{})

	// Get values for each key
	for _, key := range keys {
		value, err := s.cache.Get(ctx, key)
		if err != nil {
			continue
		}

		// Extract the counter name from the key
		counterName := key[len("realtime:counter:"):]

		// Convert value to int
		intValue, err := strconv.Atoi(value)
		if err != nil {
			continue
		}

		metrics[counterName] = intValue
	}

	return metrics, nil
}
