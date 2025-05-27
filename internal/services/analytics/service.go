package analytics

import (
	"context"

	"github.com/Caqil/vyrall/internal/config"
	"github.com/Caqil/vyrall/internal/database"
	"github.com/Caqil/vyrall/internal/utils/logger"
)

// Service provides analytics functionality
type Service struct {
	db     *database.Database
	cache  *database.RedisClient
	log    *logger.Logger
	config *config.Config

	// Sub-services
	Aggregation      *AggregationService
	ContentAnalytics *ContentAnalyticsService
	Engagement       *EngagementService
	Insights         *InsightsService
	RealTime         *RealTimeService
	Reporting        *ReportingService
	UserAnalytics    *UserAnalyticsService
}

// NewService creates a new analytics service
func NewService(db *database.Database, cache *database.RedisClient, log *logger.Logger, config *config.Config) *Service {
	service := &Service{
		db:     db,
		cache:  cache,
		log:    log,
		config: config,
	}

	// Initialize sub-services
	service.Aggregation = NewAggregationService(db, cache, log)
	service.ContentAnalytics = NewContentAnalyticsService(db, cache, log)
	service.Engagement = NewEngagementService(db, cache, log)
	service.Insights = NewInsightsService(db, cache, log)
	service.RealTime = NewRealTimeService(db, cache, log)
	service.Reporting = NewReportingService(db, cache, log)
	service.UserAnalytics = NewUserAnalyticsService(db, cache, log)

	return service
}

// TrackEvent records an analytics event
func (s *Service) TrackEvent(ctx context.Context, event *Event) error {
	// Record the event in the database
	if err := s.db.InsertOne(ctx, "analytics_events", event); err != nil {
		s.log.Error("Failed to record analytics event", "error", err)
		return err
	}

	// Process the event in real-time if enabled
	if s.config.Analytics.RealTimeProcessing {
		if err := s.RealTime.ProcessEvent(ctx, event); err != nil {
			s.log.Error("Failed to process event in real-time", "error", err)
			// Continue despite the error, as the event is already recorded
		}
	}

	return nil
}

// TrackBatchEvents records multiple analytics events
func (s *Service) TrackBatchEvents(ctx context.Context, events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	// Convert events to interface slice for batch insert
	documents := make([]interface{}, len(events))
	for i, event := range events {
		documents[i] = event
	}

	// Record the events in the database
	if err := s.db.InsertMany(ctx, "analytics_events", documents); err != nil {
		s.log.Error("Failed to record batch analytics events", "error", err)
		return err
	}

	// Process events in real-time if enabled
	if s.config.Analytics.RealTimeProcessing {
		for _, event := range events {
			if err := s.RealTime.ProcessEvent(ctx, event); err != nil {
				s.log.Error("Failed to process event in real-time", "error", err, "event_id", event.ID)
				// Continue with the next event despite the error
			}
		}
	}

	return nil
}

// GetDashboardData gets overall analytics for the dashboard
func (s *Service) GetDashboardData(ctx context.Context, period string) (*DashboardData, error) {
	// Get user metrics
	userMetrics, err := s.UserAnalytics.GetAggregatedMetrics(ctx, period)
	if err != nil {
		s.log.Error("Failed to get user metrics", "error", err)
		return nil, err
	}

	// Get content metrics
	contentMetrics, err := s.ContentAnalytics.GetAggregatedMetrics(ctx, period)
	if err != nil {
		s.log.Error("Failed to get content metrics", "error", err)
		return nil, err
	}

	// Get engagement metrics
	engagementMetrics, err := s.Engagement.GetAggregatedMetrics(ctx, period)
	if err != nil {
		s.log.Error("Failed to get engagement metrics", "error", err)
		return nil, err
	}

	// Combine all metrics into dashboard data
	dashboard := &DashboardData{
		Period:            period,
		UserMetrics:       userMetrics,
		ContentMetrics:    contentMetrics,
		EngagementMetrics: engagementMetrics,
		// Additional insights and trend analysis can be added here
	}

	return dashboard, nil
}

// RunDailyAggregation runs daily aggregation jobs
func (s *Service) RunDailyAggregation(ctx context.Context) error {
	return s.Aggregation.RunDailyAggregation(ctx)
}

// RunWeeklyAggregation runs weekly aggregation jobs
func (s *Service) RunWeeklyAggregation(ctx context.Context) error {
	return s.Aggregation.RunWeeklyAggregation(ctx)
}

// RunMonthlyAggregation runs monthly aggregation jobs
func (s *Service) RunMonthlyAggregation(ctx context.Context) error {
	return s.Aggregation.RunMonthlyAggregation(ctx)
}

// GenerateReport generates a custom analytics report
func (s *Service) GenerateReport(ctx context.Context, options *ReportOptions) (*Report, error) {
	return s.Reporting.GenerateReport(ctx, options)
}

// GetInsightsForEntity gets insights for a specific entity (user, post, etc.)
func (s *Service) GetInsightsForEntity(ctx context.Context, entityType, entityID string, period string) (*EntityInsights, error) {
	return s.Insights.GetEntityInsights(ctx, entityType, entityID, period)
}
