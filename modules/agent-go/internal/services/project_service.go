package services

import (
	"errors"
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
)

var errProjectNotFound = errors.New("project not found")

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo *repositories.ProjectRepository
}

// NewProjectService creates a new project service
func NewProjectService(projectRepo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

// ProjectDetailResponse represents a project with counts
type ProjectDetailResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	DealsCount  int64   `json:"deals_count"`
	LeadsCount  int64   `json:"leads_count"`
}

// LeadResponse represents a lead in API responses
type LeadResponse struct {
	ID       int64   `json:"id"`
	DealID   int64   `json:"deal_id"`
	Type     string  `json:"type"`
	Source   *string `json:"source"`
}

// GetProjects returns all projects for a tenant
func (s *ProjectService) GetProjects(tenantID *int64) ([]ProjectResponse, error) {
	projects, err := s.projectRepo.FindAll(tenantID)
	if err != nil {
		return nil, err
	}

	results := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		results[i] = ProjectResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Status:      p.Status,
		}
	}

	return results, nil
}

// GetProject returns a project by ID with counts
func (s *ProjectService) GetProject(id int64, tenantID *int64) (*ProjectDetailResponse, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errProjectNotFound
	}

	// Verify tenant if provided
	if tenantID != nil && project.TenantID != nil && *project.TenantID != *tenantID {
		return nil, errProjectNotFound
	}

	dealsCount, _ := s.projectRepo.GetDealsCount(id)
	leadsCount, _ := s.projectRepo.GetLeadsCount(id)

	return &ProjectDetailResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Status:      project.Status,
		DealsCount:  dealsCount,
		LeadsCount:  leadsCount,
	}, nil
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(input schemas.ProjectCreate, tenantID *int64, userID int64) (*ProjectResponse, error) {
	project := &models.Project{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		Status:      input.Status,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}

	if input.CurrentStatusID != nil {
		project.CurrentStatusID = input.CurrentStatusID
	}
	if input.StartDate != nil {
		if t, err := time.Parse("2006-01-02", *input.StartDate); err == nil {
			project.StartDate = &t
		}
	}
	if input.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *input.EndDate); err == nil {
			project.EndDate = &t
		}
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	return &ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Status:      project.Status,
	}, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(id int64, input schemas.ProjectUpdate, tenantID *int64, userID int64) (*ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errProjectNotFound
	}

	// Verify tenant if provided
	if tenantID != nil && project.TenantID != nil && *project.TenantID != *tenantID {
		return nil, errProjectNotFound
	}

	updates := buildProjectUpdates(input, userID)
	if len(updates) > 0 {
		if err := s.projectRepo.Update(id, updates); err != nil {
			return nil, err
		}
	}

	// Refetch
	project, _ = s.projectRepo.FindByID(id)
	return &ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Status:      project.Status,
	}, nil
}

func buildProjectUpdates(input schemas.ProjectUpdate, userID int64) map[string]interface{} {
	updates := map[string]interface{}{"updated_by": userID}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.CurrentStatusID != nil {
		updates["current_status_id"] = *input.CurrentStatusID
	}
	if input.StartDate != nil {
		if t, err := time.Parse("2006-01-02", *input.StartDate); err == nil {
			updates["start_date"] = t
		}
	}
	if input.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *input.EndDate); err == nil {
			updates["end_date"] = t
		}
	}
	return updates
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(id int64, tenantID *int64) error {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return err
	}
	if project == nil {
		return errProjectNotFound
	}

	// Verify tenant if provided
	if tenantID != nil && project.TenantID != nil && *project.TenantID != *tenantID {
		return errProjectNotFound
	}

	return s.projectRepo.Delete(id)
}

// GetProjectLeads returns leads for a project
func (s *ProjectService) GetProjectLeads(projectID int64, tenantID *int64) ([]LeadResponse, error) {
	leads, err := s.projectRepo.GetProjectLeads(projectID, tenantID)
	if err != nil {
		return nil, err
	}

	results := make([]LeadResponse, len(leads))
	for i, l := range leads {
		results[i] = LeadResponse{
			ID:     l.ID,
			DealID: l.DealID,
			Type:   l.Type,
			Source: l.Source,
		}
	}
	return results, nil
}

// GetProjectDeals returns deals for a project
func (s *ProjectService) GetProjectDeals(projectID int64, tenantID *int64) ([]LeadResponse, error) {
	deals, err := s.projectRepo.GetProjectDeals(projectID, tenantID)
	if err != nil {
		return nil, err
	}

	results := make([]LeadResponse, len(deals))
	for i, d := range deals {
		results[i] = LeadResponse{
			ID:     d.ID,
			DealID: d.DealID,
			Type:   d.Type,
			Source: d.Source,
		}
	}
	return results, nil
}
