package services

import (
	"github.com/pina-colada-co/agent-go/internal/repositories"
)

// UsageService handles usage analytics business logic
type UsageService struct {
	usageRepo *repositories.UsageRepository
	userRepo  *repositories.UserRepository
}

// NewUsageService creates a new usage service
func NewUsageService(usageRepo *repositories.UsageRepository, userRepo *repositories.UserRepository) *UsageService {
	return &UsageService{usageRepo: usageRepo, userRepo: userRepo}
}

// UsageTotals represents aggregated usage totals
type UsageTotals struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func aggregateUsage(records []repositories.UsageRecord) UsageTotals {
	totals := UsageTotals{}
	for _, r := range records {
		totals.InputTokens += r.InputTokens
		totals.OutputTokens += r.OutputTokens
		totals.TotalTokens += r.TotalTokens
	}
	return totals
}

// GetUserUsage returns aggregated usage for a user
func (s *UsageService) GetUserUsage(userID int64, period string) (*UsageTotals, error) {
	records, err := s.usageRepo.GetUserUsage(userID, period)
	if err != nil {
		return nil, err
	}
	totals := aggregateUsage(records)
	return &totals, nil
}

// GetTenantUsage returns aggregated usage for a tenant
func (s *UsageService) GetTenantUsage(tenantID int64, period string) (*UsageTotals, error) {
	records, err := s.usageRepo.GetTenantUsage(tenantID, period)
	if err != nil {
		return nil, err
	}
	totals := aggregateUsage(records)
	return &totals, nil
}

// GetUsageTimeseries returns usage data grouped by date for charting
func (s *UsageService) GetUsageTimeseries(tenantID int64, period string, userID *int64) ([]repositories.TimeseriesRecord, error) {
	return s.usageRepo.GetUsageTimeseries(tenantID, period, userID)
}

// DeveloperAnalyticsResponse represents developer analytics data
type DeveloperAnalyticsResponse struct {
	GroupBy string      `json:"group_by"`
	Data    interface{} `json:"data"`
}

// GetDeveloperAnalytics returns developer analytics data
func (s *UsageService) GetDeveloperAnalytics(userID, tenantID int64, period, groupBy string) (*DeveloperAnalyticsResponse, error) {
	// Check if user has developer role
	hasDeveloperRole, err := s.userRepo.HasRole(userID, "developer")
	if err != nil {
		return nil, err
	}
	if !hasDeveloperRole {
		return &DeveloperAnalyticsResponse{
			GroupBy: groupBy,
			Data:    []interface{}{},
		}, nil
	}

	var data interface{}
	if groupBy == "model" {
		data, err = s.usageRepo.GetUsageByModel(tenantID, period)
	}
	if groupBy != "model" {
		data, err = s.usageRepo.GetUsageByNode(tenantID, period)
	}
	if err != nil {
		return nil, err
	}

	return &DeveloperAnalyticsResponse{
		GroupBy: groupBy,
		Data:    data,
	}, nil
}

// DeveloperAccessResponse represents developer access check result
type DeveloperAccessResponse struct {
	HasDeveloperAccess bool `json:"has_developer_access"`
}

// CheckDeveloperAccess checks if user has developer role
func (s *UsageService) CheckDeveloperAccess(userID int64) (*DeveloperAccessResponse, error) {
	hasRole, err := s.userRepo.HasRole(userID, "developer")
	if err != nil {
		return nil, err
	}
	return &DeveloperAccessResponse{HasDeveloperAccess: hasRole}, nil
}
