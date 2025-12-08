package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// ProjectRepository handles project data access
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// FindAll returns all projects for a tenant
func (r *ProjectRepository) FindAll(tenantID *int64) ([]models.Project, error) {
	var projects []models.Project

	query := r.db.Model(&models.Project{}).Order("name ASC")

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	err := query.Find(&projects).Error
	return projects, err
}

// FindByID returns a project by ID
func (r *ProjectRepository) FindByID(id int64) (*models.Project, error) {
	var project models.Project
	err := r.db.First(&project, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}
