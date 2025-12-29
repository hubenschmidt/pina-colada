package services

import (
	"errors"

	"agent/internal/lib"
	"agent/internal/repositories"
	"agent/internal/schemas"
	"agent/internal/serializers"
)

// ErrProjectNotFound is returned when a project is not found
var ErrProjectNotFound = errors.New("project not found")

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo *repositories.ProjectRepository
}

// NewProjectService creates a new project service
func NewProjectService(projectRepo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

// GetProjects returns all projects for a tenant
func (s *ProjectService) GetProjects(tenantID *int64) ([]serializers.ProjectResponse, error) {
	projects, err := s.projectRepo.FindAll(tenantID)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.ProjectResponse, len(projects))
	for i, p := range projects {
		results[i] = serializers.ProjectToResponse(&p)
	}

	return results, nil
}

// GetProject returns a project by ID with counts
func (s *ProjectService) GetProject(id int64, tenantID *int64) (*serializers.ProjectDetailResponse, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	// Verify tenant if provided
	if tenantID != nil && project.TenantID != nil && *project.TenantID != *tenantID {
		return nil, ErrProjectNotFound
	}

	dealsCount, _ := s.projectRepo.GetDealsCount(id)
	leadsCount, _ := s.projectRepo.GetLeadsCount(id)

	resp := serializers.ProjectToDetailResponse(project, dealsCount, leadsCount)
	return &resp, nil
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(input schemas.ProjectCreate, tenantID *int64, userID int64) (*serializers.ProjectResponse, error) {
	repoInput := repositories.ProjectCreateInput{
		TenantID:        tenantID,
		Name:            input.Name,
		Description:     input.Description,
		Status:          input.Status,
		CurrentStatusID: input.CurrentStatusID,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	}

	project, err := s.projectRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}

	resp := serializers.ProjectToResponse(project)
	return &resp, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(id int64, input schemas.ProjectUpdate, tenantID *int64, userID int64) (*serializers.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	// Verify tenant if provided
	if tenantID != nil && project.TenantID != nil && *project.TenantID != *tenantID {
		return nil, ErrProjectNotFound
	}

	updates := buildProjectUpdates(input, userID)
	var updateErr error
	if len(updates) > 0 {
		updateErr = s.projectRepo.Update(id, updates)
	}
	if updateErr != nil {
		return nil, updateErr
	}

	// Refetch
	project, _ = s.projectRepo.FindByID(id)
	resp := serializers.ProjectToResponse(project)
	return &resp, nil
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
	if t := lib.ParseDateString(input.StartDate); t != nil {
		updates["start_date"] = *t
	}
	if t := lib.ParseDateString(input.EndDate); t != nil {
		updates["end_date"] = *t
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
		return ErrProjectNotFound
	}

	// Verify tenant if provided
	if tenantID != nil && project.TenantID != nil && *project.TenantID != *tenantID {
		return ErrProjectNotFound
	}

	return s.projectRepo.Delete(id)
}

// GetProjectLeads returns leads for a project
func (s *ProjectService) GetProjectLeads(projectID int64, tenantID *int64) ([]serializers.ProjectLeadResponse, error) {
	leads, err := s.projectRepo.GetProjectLeads(projectID, tenantID)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.ProjectLeadResponse, len(leads))
	for i, l := range leads {
		results[i] = serializers.ProjectLeadResponse{
			ID:     l.ID,
			DealID: l.DealID,
			Type:   l.Type,
			Source: l.Source,
		}
	}
	return results, nil
}

// GetProjectDeals returns deals for a project
func (s *ProjectService) GetProjectDeals(projectID int64, tenantID *int64) ([]serializers.ProjectLeadResponse, error) {
	deals, err := s.projectRepo.GetProjectDeals(projectID, tenantID)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.ProjectLeadResponse, len(deals))
	for i, d := range deals {
		results[i] = serializers.ProjectLeadResponse{
			ID:     d.ID,
			DealID: d.DealID,
			Type:   d.Type,
			Source: d.Source,
		}
	}
	return results, nil
}
