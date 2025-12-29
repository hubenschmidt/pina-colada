package serializers

import (
	"time"

	"agent/internal/models"
)

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

// ProjectDetailResponse represents a project with counts
type ProjectDetailResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	DealsCount  int64   `json:"deals_count"`
	LeadsCount  int64   `json:"leads_count"`
}

// ProjectLeadResponse represents a lead in project API responses
type ProjectLeadResponse struct {
	ID     int64   `json:"id"`
	DealID int64   `json:"deal_id"`
	Type   string  `json:"type"`
	Source *string `json:"source"`
}

func formatProjectDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// ProjectToResponse converts a Project model to response
func ProjectToResponse(p *models.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
		StartDate:   formatProjectDate(p.StartDate),
		EndDate:     formatProjectDate(p.EndDate),
	}
}

// ProjectToDetailResponse converts a Project model to detail response
func ProjectToDetailResponse(p *models.Project, dealsCount, leadsCount int64) ProjectDetailResponse {
	return ProjectDetailResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
		StartDate:   formatProjectDate(p.StartDate),
		EndDate:     formatProjectDate(p.EndDate),
		DealsCount:  dealsCount,
		LeadsCount:  leadsCount,
	}
}
