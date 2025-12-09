package services

import (
	"encoding/json"
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"gorm.io/datatypes"
)

// ReportService handles report business logic
type ReportService struct {
	reportRepo *repositories.ReportRepository
}

// NewReportService creates a new report service
func NewReportService(reportRepo *repositories.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

// SavedReportResponse represents a saved report in API responses
type SavedReportResponse struct {
	ID              int64          `json:"id"`
	TenantID        int64          `json:"tenant_id"`
	Name            string         `json:"name"`
	Description     *string        `json:"description"`
	QueryDefinition map[string]any `json:"query_definition"`
	CreatedBy       *int64         `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	ProjectIDs      []int64        `json:"project_ids"`
}

// SavedReportsListResponse represents paginated saved reports
type SavedReportsListResponse struct {
	Items       []SavedReportResponse `json:"items"`
	Total       int64                 `json:"total"`
	CurrentPage int                   `json:"currentPage"`
	TotalPages  int                   `json:"totalPages"`
	PageSize    int                   `json:"pageSize"`
}

// EntityFieldsResponse represents available fields for an entity
type EntityFieldsResponse struct {
	Base           []string          `json:"base"`
	Joins          []string          `json:"joins"`
	AvailableJoins []AvailableJoin   `json:"available_joins"`
}

// AvailableJoin represents an available join for an entity
type AvailableJoin struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

// CustomQueryResponse represents the result of a custom query
type CustomQueryResponse struct {
	Data   []map[string]interface{} `json:"data"`
	Total  int64                    `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}

// CannedReportResponse represents a canned report result
type CannedReportResponse map[string]interface{}

// Entity field definitions
var entityFields = map[string][]string{
	"organizations": {"id", "name", "website", "phone", "employee_count", "description", "founding_year", "headquarters_city", "headquarters_state", "headquarters_country", "company_type", "linkedin_url", "crunchbase_url", "created_at", "updated_at"},
	"individuals":   {"id", "first_name", "last_name", "email", "phone", "title", "department", "seniority_level", "linkedin_url", "twitter_url", "github_url", "bio", "is_decision_maker", "created_at", "updated_at"},
	"contacts":      {"id", "first_name", "last_name", "email", "phone", "title", "department", "role", "is_primary", "notes", "created_at", "updated_at"},
	"leads":         {"id", "title", "description", "source", "type", "created_at", "updated_at"},
	"notes":         {"id", "entity_type", "entity_id", "content", "created_at", "updated_at"},
}

var entityJoins = map[string]map[string]AvailableJoin{
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

func reportToResponse(report *models.SavedReport, projectIDs []int64) SavedReportResponse {
	var queryDef map[string]any
	if report.QueryDefinition != nil {
		json.Unmarshal(report.QueryDefinition, &queryDef)
	}

	return SavedReportResponse{
		ID:              report.ID,
		TenantID:        report.TenantID,
		Name:            report.Name,
		Description:     report.Description,
		QueryDefinition: queryDef,
		CreatedBy:       report.CreatedBy,
		CreatedAt:       report.CreatedAt,
		UpdatedAt:       report.UpdatedAt,
		ProjectIDs:      projectIDs,
	}
}

// GetEntityFields returns available fields for an entity
func (s *ReportService) GetEntityFields(entity string) (*EntityFieldsResponse, error) {
	baseFields, ok := entityFields[entity]
	if !ok {
		return nil, nil
	}

	joins := entityJoins[entity]
	var joinFields []string
	var availableJoins []AvailableJoin

	for _, join := range joins {
		joinFields = append(joinFields, join.Fields...)
		availableJoins = append(availableJoins, join)
	}

	return &EntityFieldsResponse{
		Base:           baseFields,
		Joins:          joinFields,
		AvailableJoins: availableJoins,
	}, nil
}

// GetSavedReports returns saved reports for a tenant
func (s *ReportService) GetSavedReports(tenantID int64, projectID *int64, includeGlobal bool, search string, page, limit int, sortBy, order string) (*SavedReportsListResponse, error) {
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

	items := make([]SavedReportResponse, len(reports))
	for i, r := range reports {
		projectIDs, _ := s.reportRepo.GetProjectIDsForReport(r.ID)
		items[i] = reportToResponse(&r, projectIDs)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &SavedReportsListResponse{
		Items:       items,
		Total:       total,
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    limit,
	}, nil
}

// GetSavedReport returns a saved report by ID
func (s *ReportService) GetSavedReport(reportID int64, tenantID int64) (*SavedReportResponse, error) {
	report, err := s.reportRepo.FindSavedReportByID(reportID, tenantID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, nil
	}

	projectIDs, _ := s.reportRepo.GetProjectIDsForReport(reportID)
	resp := reportToResponse(report, projectIDs)
	return &resp, nil
}

// CreateSavedReport creates a new saved report
func (s *ReportService) CreateSavedReport(tenantID int64, userID *int64, input schemas.SavedReportCreate) (*SavedReportResponse, error) {
	queryDefBytes, err := json.Marshal(input.QueryDefinition)
	if err != nil {
		return nil, err
	}

	report := &models.SavedReport{
		TenantID:        tenantID,
		Name:            input.Name,
		Description:     input.Description,
		QueryDefinition: datatypes.JSON(queryDefBytes),
		CreatedBy:       userID,
	}

	if err := s.reportRepo.CreateSavedReport(report); err != nil {
		return nil, err
	}

	if len(input.ProjectIDs) > 0 {
		s.reportRepo.SetProjectsForReport(report.ID, input.ProjectIDs)
	}

	return s.GetSavedReport(report.ID, tenantID)
}

// UpdateSavedReport updates a saved report
func (s *ReportService) UpdateSavedReport(reportID int64, tenantID int64, input schemas.SavedReportUpdate) (*SavedReportResponse, error) {
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
		updates["query_definition"] = datatypes.JSON(queryDefBytes)
	}

	if len(updates) > 0 {
		if err := s.reportRepo.UpdateSavedReport(reportID, updates); err != nil {
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
func (s *ReportService) PreviewCustomReport(tenantID int64, query schemas.ReportQueryRequest) (*CustomQueryResponse, error) {
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

	return &CustomQueryResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: query.Offset,
	}, nil
}

// RunCustomReport executes a custom report with full results
func (s *ReportService) RunCustomReport(tenantID int64, query schemas.ReportQueryRequest) (*CustomQueryResponse, error) {
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

	return &CustomQueryResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: query.Offset,
	}, nil
}

// ExportCustomReport exports a custom report (returns data for Excel generation)
func (s *ReportService) ExportCustomReport(tenantID int64, query schemas.ReportQueryRequest) (*CustomQueryResponse, error) {
	query.Limit = 10000
	query.Offset = 0
	return s.RunCustomReport(tenantID, query)
}

// GetLeadPipelineReport returns the lead pipeline canned report
func (s *ReportService) GetLeadPipelineReport(tenantID int64, dateFrom, dateTo *string, projectID *int64) (CannedReportResponse, error) {
	data, err := s.reportRepo.GetLeadPipelineData(tenantID, dateFrom, dateTo, projectID)
	if err != nil {
		return nil, err
	}

	// Group by type
	typeCounts := make(map[string]int)
	for _, row := range data {
		leadType, _ := row["type"].(string)
		if leadType == "" {
			leadType = "unknown"
		}
		typeCounts[leadType]++
	}

	return CannedReportResponse{
		"leads":       data,
		"type_counts": typeCounts,
		"total":       len(data),
	}, nil
}

// GetAccountOverviewReport returns the account overview canned report
func (s *ReportService) GetAccountOverviewReport(tenantID int64) (CannedReportResponse, error) {
	data, err := s.reportRepo.GetAccountOverviewData(tenantID)
	if err != nil {
		return nil, err
	}
	return CannedReportResponse(data), nil
}

// GetContactCoverageReport returns the contact coverage canned report
func (s *ReportService) GetContactCoverageReport(tenantID int64) (CannedReportResponse, error) {
	data, err := s.reportRepo.GetContactCoverageData(tenantID)
	if err != nil {
		return nil, err
	}
	return CannedReportResponse{
		"organizations": data,
		"total":         len(data),
	}, nil
}

// GetNotesActivityReport returns the notes activity canned report
func (s *ReportService) GetNotesActivityReport(tenantID int64, projectID *int64) (CannedReportResponse, error) {
	data, err := s.reportRepo.GetNotesActivityData(tenantID, projectID)
	if err != nil {
		return nil, err
	}
	return CannedReportResponse{
		"notes": data,
		"total": len(data),
	}, nil
}

// GetUserAuditReport returns the user audit canned report
func (s *ReportService) GetUserAuditReport(tenantID int64, userID *int64) (CannedReportResponse, error) {
	data, err := s.reportRepo.GetUserAuditData(tenantID, userID)
	if err != nil {
		return nil, err
	}
	return CannedReportResponse(data), nil
}
