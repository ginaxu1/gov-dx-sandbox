package services

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type AdminService struct {
	consumerService *ConsumerService
	providerService *ProviderService
	mutex           sync.RWMutex
}

func NewAdminService() *AdminService {
	return &AdminService{
		consumerService: NewConsumerService(),
		providerService: NewProviderService(),
	}
}

// NewAdminServiceWithServices creates an admin service with existing services
func NewAdminServiceWithServices(consumerService *ConsumerService, providerService *ProviderService) *AdminService {
	return &AdminService{
		consumerService: consumerService,
		providerService: providerService,
	}
}

// GetConsumerService returns the consumer service for testing
func (s *AdminService) GetConsumerService() *ConsumerService {
	return s.consumerService
}

// GetProviderService returns the provider service for testing
func (s *AdminService) GetProviderService() *ProviderService {
	return s.providerService
}

// GetMetrics returns basic system metrics
func (s *AdminService) GetMetrics() (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get counts from services
	applications, err := s.consumerService.GetAllApplications()
	if err != nil {
		slog.Error("Failed to get applications for metrics", "error", err)
		return nil, err
	}

	submissions, err := s.providerService.GetAllProviderSubmissions()
	if err != nil {
		slog.Error("Failed to get provider submissions for metrics", "error", err)
		return nil, err
	}

	profiles, err := s.providerService.GetAllProviderProfiles()
	if err != nil {
		slog.Error("Failed to get provider profiles for metrics", "error", err)
		return nil, err
	}

	schemas, err := s.providerService.GetAllProviderSchemas()
	if err != nil {
		slog.Error("Failed to get provider schemas for metrics", "error", err)
		return nil, err
	}

	metrics := map[string]interface{}{
		"total_consumer_apps":        len(applications),
		"total_provider_submissions": len(submissions),
		"total_providers":            len(profiles),
		"total_schemas":              len(schemas),
	}

	return metrics, nil
}

// GetRecentActivity returns recent system activity
func (s *AdminService) GetRecentActivity() ([]map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get data from services
	applications, err := s.consumerService.GetAllApplications()
	if err != nil {
		slog.Error("Failed to get applications for recent activity", "error", err)
		return nil, err
	}

	submissions, err := s.providerService.GetAllProviderSubmissions()
	if err != nil {
		slog.Error("Failed to get provider submissions for recent activity", "error", err)
		return nil, err
	}

	profiles, err := s.providerService.GetAllProviderProfiles()
	if err != nil {
		slog.Error("Failed to get provider profiles for recent activity", "error", err)
		return nil, err
	}

	schemas, err := s.providerService.GetAllProviderSchemas()
	if err != nil {
		slog.Error("Failed to get provider schemas for recent activity", "error", err)
		return nil, err
	}

	// Generate recent activity
	recentActivity := s.generateRecentActivity(applications, submissions, profiles, schemas)

	return recentActivity, nil
}

// GetStatistics returns detailed statistics by resource type
func (s *AdminService) GetStatistics() (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get data from services
	applications, err := s.consumerService.GetAllApplications()
	if err != nil {
		slog.Error("Failed to get applications for statistics", "error", err)
		return nil, err
	}

	submissions, err := s.providerService.GetAllProviderSubmissions()
	if err != nil {
		slog.Error("Failed to get provider submissions for statistics", "error", err)
		return nil, err
	}

	schemas, err := s.providerService.GetAllProviderSchemas()
	if err != nil {
		slog.Error("Failed to get provider schemas for statistics", "error", err)
		return nil, err
	}

	// Count by status
	applicationStats := s.countApplicationsByStatus(applications)
	submissionStats := s.countSubmissionsByStatus(submissions)
	schemaStats := s.countSchemasByStatus(schemas)

	statistics := map[string]interface{}{
		"consumer-apps":        applicationStats,
		"provider-submissions": submissionStats,
		"provider-schemas":     schemaStats,
	}

	return statistics, nil
}

// countApplicationsByStatus counts applications by their status
func (s *AdminService) countApplicationsByStatus(applications []*models.Application) map[string]int {
	stats := make(map[string]int)
	for _, app := range applications {
		stats[string(app.Status)]++
	}
	return stats
}

// countSubmissionsByStatus counts submissions by their status
func (s *AdminService) countSubmissionsByStatus(submissions []*models.ProviderSubmission) map[string]int {
	stats := make(map[string]int)
	for _, sub := range submissions {
		stats[string(sub.Status)]++
	}
	return stats
}

// countSchemasByStatus counts schemas by their status
func (s *AdminService) countSchemasByStatus(schemas []*models.ProviderSchema) map[string]int {
	stats := make(map[string]int)
	for _, schema := range schemas {
		stats[string(schema.Status)]++
	}
	return stats
}

// ActivityItem represents a single activity item for the dashboard
type ActivityItem struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	ID          string    `json:"id"`
}

// generateRecentActivity creates a list of recent activities from all data sources
func (s *AdminService) generateRecentActivity(applications []*models.Application, submissions []*models.ProviderSubmission, profiles []*models.ProviderProfile, schemas []*models.ProviderSchema) []map[string]interface{} {
	var activities []ActivityItem

	// Add application activities
	for _, app := range applications {
		activities = append(activities, ActivityItem{
			Type:        "application_created",
			Description: "New consumer application submitted",
			Timestamp:   time.Now(), // In a real system, this would be app.CreatedAt
			ID:          app.SubmissionID,
		})
	}

	// Add submission activities
	for _, sub := range submissions {
		activities = append(activities, ActivityItem{
			Type:        "submission_created",
			Description: "New provider submission submitted",
			Timestamp:   sub.CreatedAt,
			ID:          sub.SubmissionID,
		})
	}

	// Add profile activities (when submissions are approved)
	for _, profile := range profiles {
		activities = append(activities, ActivityItem{
			Type:        "provider_approved",
			Description: "Provider submission approved",
			Timestamp:   profile.ApprovedAt,
			ID:          profile.ProviderID,
		})
	}

	// Add schema activities
	for _, schema := range schemas {
		activities = append(activities, ActivityItem{
			Type:        "schema_submitted",
			Description: "Provider schema submitted",
			Timestamp:   time.Now(), // In a real system, this would be schema.CreatedAt
			ID:          schema.SubmissionID,
		})
	}

	// Sort by timestamp (most recent first)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	// Convert to map format and limit to 10 most recent
	limit := 10
	if len(activities) < limit {
		limit = len(activities)
	}

	result := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		result[i] = map[string]interface{}{
			"type":        activities[i].Type,
			"description": activities[i].Description,
			"timestamp":   activities[i].Timestamp.Format(time.RFC3339),
			"id":          activities[i].ID,
		}
	}

	return result
}
