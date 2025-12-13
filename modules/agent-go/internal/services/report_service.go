package services

import (
	"encoding/json"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

// ReportService handles report business logic
type ReportService struct {
	reportRepo *repositories.ReportRepository
}

// NewReportService creates a new report service
func NewReportService(reportRepo *repositories.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

// Entity field definitions
var entityFields = map[string][]string{
	"organizations": {"id", "name", "website", "phone", "employee_count", "description", "founding_year", "headquarters_city", "headquarters_state", "headquarters_country", "company_type", "linkedin_url", "crunchbase_url", "created_at", "updated_at"},
	"individuals":   {"id", "first_name", "last_name", "email", "phone", "title", "department", "seniority_level", "linkedin_url", "twitter_url", "github_url", "bio", "is_decision_maker", "created_at", "updated_at"},
	"contacts":      {"id", "first_name", "last_name", "email", "phone", "title", "department", "role", "is_primary", "notes", "created_at", "updated_at"},
	"leads":         {"id", "source", "type", "deal_id", "current_status_id", "created_at", "updated_at"},
	"notes":         {"id", "entity_type", "entity_id", "content", "created_at", "updated_at"},
}

var entityJoins = map[string]map[string]serializers.AvailableJoin{
	"organizations": {
		"account":              {Name: "account", Fields: []string{"account.name"}},
		"employee_count_range": {Name: "employee_count_range", Fields: []string{"employee_count_range.label"}},
		"funding_stage":        {Name: "funding_stage", Fields: []string{"funding_stage.name"}},
		"revenue_range":        {Name: "revenue_range", Fields: []string{"revenue_range.label"}},
	},
	"individuals": {
		"account": {Name: "account", Fields: []string{"account.name"}},
	},
	"contacts": {
		"organizations": {Name: "organizations", Fields: []string{"organization.name"}},
		"individuals":   {Name: "individuals", Fields: []string{"individual.first_name", "individual.last_name"}},
	},
	"leads": {
		"account": {Name: "account", Fields: []string{"account.name"}},
	},
	"notes": {},
}

// GetEntityFields returns available fields for an entity
func (s *ReportService) GetEntityFields(entity string) (*serializers.EntityFieldsResponse, error) {
	baseFields, ok := entityFields[entity]
	if !ok {
		return nil, nil
	}

	joins := entityJoins[entity]
	var joinFields []string
	var availableJoins []serializers.AvailableJoin

	for _, join := range joins {
		joinFields = append(joinFields, join.Fields...)
		availableJoins = append(availableJoins, join)
	}

	return &serializers.EntityFieldsResponse{
		Base:           baseFields,
		Joins:          joinFields,
		AvailableJoins: availableJoins,
	}, nil
}

// GetSavedReports returns saved reports for a tenant
func (s *ReportService) GetSavedReports(tenantID int64, projectID *int64, includeGlobal bool, search string, page, limit int, sortBy, order string) (*serializers.SavedReportsListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}

	reports, total, err := s.reportRepo.FindSavedReports(tenantID, projectID, includeGlobal, search, page, limit, sortBy, order)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.SavedReportResponse, len(reports))
	for i, r := range reports {
		projectIDs, _ := s.reportRepo.GetProjectIDsForReport(r.ID)
		items[i] = serializers.SavedReportToResponse(&r, projectIDs)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &serializers.SavedReportsListResponse{
		Items:       items,
		Total:       total,
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    limit,
	}, nil
}

// GetSavedReport returns a saved report by ID
func (s *ReportService) GetSavedReport(reportID int64, tenantID int64) (*serializers.SavedReportResponse, error) {
	report, err := s.reportRepo.FindSavedReportByID(reportID, tenantID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, nil
	}

	projectIDs, _ := s.reportRepo.GetProjectIDsForReport(reportID)
	resp := serializers.SavedReportToResponse(report, projectIDs)
	return &resp, nil
}

// CreateSavedReport creates a new saved report
func (s *ReportService) CreateSavedReport(tenantID int64, userID *int64, input schemas.SavedReportCreate) (*serializers.SavedReportResponse, error) {
	queryDefBytes, err := json.Marshal(input.QueryDefinition)
	if err != nil {
		return nil, err
	}

	repoInput := repositories.SavedReportCreateInput{
		TenantID:        tenantID,
		Name:            input.Name,
		Description:     input.Description,
		QueryDefinition: queryDefBytes,
		CreatedBy:       userID,
	}

	reportID, err := s.reportRepo.CreateSavedReport(repoInput)
	if err != nil {
		return nil, err
	}

	if len(input.ProjectIDs) > 0 {
		s.reportRepo.SetProjectsForReport(reportID, input.ProjectIDs)
	}

	return s.GetSavedReport(reportID, tenantID)
}

// UpdateSavedReport updates a saved report
func (s *ReportService) UpdateSavedReport(reportID int64, tenantID int64, input schemas.SavedReportUpdate) (*serializers.SavedReportResponse, error) {
	report, err := s.reportRepo.FindSavedReportByID(reportID, tenantID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, nil
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.QueryDefinition != nil {
		queryDefBytes, err := json.Marshal(input.QueryDefinition)
		if err != nil {
			return nil, err
		}
		updates["query_definition"] = queryDefBytes
	}

	if len(updates) > 0 {
		err := s.reportRepo.UpdateSavedReport(reportID, updates)
		if err != nil {
			return nil, err
		}
	}

	if input.ProjectIDs != nil {
		s.reportRepo.SetProjectsForReport(reportID, input.ProjectIDs)
	}

	return s.GetSavedReport(reportID, tenantID)
}

// DeleteSavedReport deletes a saved report
func (s *ReportService) DeleteSavedReport(reportID int64, tenantID int64) (bool, error) {
	report, err := s.reportRepo.FindSavedReportByID(reportID, tenantID)
	if err != nil {
		return false, err
	}
	if report == nil {
		return false, nil
	}

	if err := s.reportRepo.DeleteSavedReport(reportID); err != nil {
		return false, err
	}
	return true, nil
}

// PreviewCustomReport executes a custom report with limited rows
func (s *ReportService) PreviewCustomReport(tenantID int64, query schemas.ReportQueryRequest) (*serializers.CustomQueryResponse, error) {
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	filters := make([]map[string]interface{}, len(query.Filters))
	for i, f := range query.Filters {
		filters[i] = map[string]interface{}{
			"field":    f.Field,
			"operator": f.Operator,
			"value":    f.Value,
		}
	}

	data, total, err := s.reportRepo.ExecuteCustomQuery(tenantID, query.PrimaryEntity, filters, limit, query.Offset, query.ProjectID)
	if err != nil {
		return nil, err
	}

	return &serializers.CustomQueryResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: query.Offset,
	}, nil
}

// RunCustomReport executes a custom report with full results
func (s *ReportService) RunCustomReport(tenantID int64, query schemas.ReportQueryRequest) (*serializers.CustomQueryResponse, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 1000
	}

	filters := make([]map[string]interface{}, len(query.Filters))
	for i, f := range query.Filters {
		filters[i] = map[string]interface{}{
			"field":    f.Field,
			"operator": f.Operator,
			"value":    f.Value,
		}
	}

	data, total, err := s.reportRepo.ExecuteCustomQuery(tenantID, query.PrimaryEntity, filters, limit, query.Offset, query.ProjectID)
	if err != nil {
		return nil, err
	}

	return &serializers.CustomQueryResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: query.Offset,
	}, nil
}

// ExportCustomReport exports a custom report (returns data for Excel generation)
func (s *ReportService) ExportCustomReport(tenantID int64, query schemas.ReportQueryRequest) (*serializers.CustomQueryResponse, error) {
	query.Limit = 10000
	query.Offset = 0
	return s.RunCustomReport(tenantID, query)
}

// GetLeadPipelineReport returns the lead pipeline canned report
func (s *ReportService) GetLeadPipelineReport(tenantID int64, dateFrom, dateTo *string, projectID *int64) (serializers.CannedReportResponse, error) {
	data, err := s.reportRepo.GetLeadPipelineData(tenantID, dateFrom, dateTo, projectID)
	if err != nil {
		return nil, err
	}

	// Group by type and source
	byType := make(map[string]int)
	bySource := make(map[string]int)
	for _, row := range data {
		leadType, _ := row["type"].(string)
		if leadType == "" {
			leadType = "Unknown"
		}
		byType[leadType]++

		source, _ := row["source"].(string)
		if source == "" {
			source = "Unknown"
		}
		bySource[source]++
	}

	return serializers.CannedReportResponse{
		"total_leads": len(data),
		"by_type":     byType,
		"by_source":   bySource,
	}, nil
}

// GetAccountOverviewReport returns the account overview canned report
func (s *ReportService) GetAccountOverviewReport(tenantID int64) (serializers.CannedReportResponse, error) {
	data, err := s.reportRepo.GetAccountOverviewData(tenantID)
	if err != nil {
		return nil, err
	}
	return serializers.CannedReportResponse(data), nil
}

// GetContactCoverageReport returns the contact coverage canned report
func (s *ReportService) GetContactCoverageReport(tenantID int64) (serializers.CannedReportResponse, error) {
	data, err := s.reportRepo.GetContactCoverageData(tenantID)
	if err != nil {
		return nil, err
	}
	return serializers.CannedReportResponse(data), nil
}

// GetNotesActivityReport returns the notes activity canned report
func (s *ReportService) GetNotesActivityReport(tenantID int64, projectID *int64) (serializers.CannedReportResponse, error) {
	data, err := s.reportRepo.GetNotesActivityData(tenantID, projectID)
	if err != nil {
		return nil, err
	}
	return serializers.CannedReportResponse(data), nil
}

// GetUserAuditReport returns the user audit canned report
func (s *ReportService) GetUserAuditReport(tenantID int64, userID *int64) (serializers.CannedReportResponse, error) {
	data, err := s.reportRepo.GetUserAuditData(tenantID, userID)
	if err != nil {
		return nil, err
	}
	return serializers.CannedReportResponse(data), nil
}
