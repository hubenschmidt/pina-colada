package serializers

import (
	"encoding/json"
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
)

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
	ProjectNames    []string       `json:"project_names"`
	IsGlobal        bool           `json:"is_global"`
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
	Base           []string        `json:"base"`
	Joins          []string        `json:"joins"`
	AvailableJoins []AvailableJoin `json:"available_joins"`
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

// SavedReportToResponse converts a SavedReport model to response
func SavedReportToResponse(report *models.SavedReport, projectIDs []int64, projectNames []string) SavedReportResponse {
	var queryDef map[string]any
	if report.QueryDefinition != nil {
		json.Unmarshal(report.QueryDefinition, &queryDef)
	}

	if projectIDs == nil {
		projectIDs = make([]int64, 0)
	}
	if projectNames == nil {
		projectNames = make([]string, 0)
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
		ProjectNames:    projectNames,
		IsGlobal:        len(projectIDs) == 0,
	}
}
