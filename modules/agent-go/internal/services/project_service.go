package services

import (
	"github.com/pina-colada-co/agent-go/internal/repositories"
)

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
