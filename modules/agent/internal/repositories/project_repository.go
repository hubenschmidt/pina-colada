package repositories

import (
	"errors"

	apperrors "agent/internal/errors"
	"agent/internal/lib"
	"agent/internal/models"
	"gorm.io/gorm"
)

// ProjectRepository handles project data access
type ProjectRepository struct {
	db *gorm.DB
}

// ProjectCreateInput represents input for creating a project
type ProjectCreateInput struct {
	TenantID        *int64
	Name            string
	Description     *string
	Status          *string
	CurrentStatusID *int64
	StartDate       *string
	EndDate         *string
	CreatedBy       int64
	UpdatedBy       int64
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Create creates a new project
func (r *ProjectRepository) Create(input ProjectCreateInput) (*models.Project, error) {
	project := &models.Project{
		TenantID:        input.TenantID,
		Name:            input.Name,
		Description:     input.Description,
		Status:          input.Status,
		CurrentStatusID: input.CurrentStatusID,
		CreatedBy:       input.CreatedBy,
		UpdatedBy:       input.UpdatedBy,
	}

	project.StartDate = lib.ParseDateString(input.StartDate)
	project.EndDate = lib.ParseDateString(input.EndDate)

	if err := r.db.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

// Update updates a project
func (r *ProjectRepository) Update(id int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Project{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a project
func (r *ProjectRepository) Delete(id int64) error {
	return r.db.Delete(&models.Project{}, id).Error
}

// GetDealsCount returns the count of deals for a project
func (r *ProjectRepository) GetDealsCount(projectID int64) (int64, error) {
	var count int64
	err := r.db.Table(`"Lead_Project"`).
		Joins(`JOIN "Lead" ON "Lead".id = "Lead_Project".lead_id`).
		Where(`"Lead_Project".project_id = ? AND "Lead".lead_type = ?`, projectID, "Deal").
		Count(&count).Error
	return count, err
}

// GetLeadsCount returns the count of leads (non-deals) for a project
func (r *ProjectRepository) GetLeadsCount(projectID int64) (int64, error) {
	var count int64
	err := r.db.Table(`"Lead_Project"`).
		Joins(`JOIN "Lead" ON "Lead".id = "Lead_Project".lead_id`).
		Where(`"Lead_Project".project_id = ? AND "Lead".lead_type != ?`, projectID, "Deal").
		Count(&count).Error
	return count, err
}

// GetProjectLeads returns leads for a project (excluding deals)
func (r *ProjectRepository) GetProjectLeads(projectID int64, tenantID *int64) ([]models.Lead, error) {
	var leads []models.Lead
	query := r.db.Table(`"Lead"`).
		Joins(`JOIN "Lead_Project" ON "Lead".id = "Lead_Project".lead_id`).
		Where(`"Lead_Project".project_id = ? AND "Lead".lead_type != ?`, projectID, "Deal")

	if tenantID != nil {
		query = query.Where(`"Lead".tenant_id = ?`, *tenantID)
	}

	err := query.Find(&leads).Error
	return leads, err
}

// GetProjectDeals returns deals for a project
func (r *ProjectRepository) GetProjectDeals(projectID int64, tenantID *int64) ([]models.Lead, error) {
	var deals []models.Lead
	query := r.db.Table(`"Lead"`).
		Joins(`JOIN "Lead_Project" ON "Lead".id = "Lead_Project".lead_id`).
		Where(`"Lead_Project".project_id = ? AND "Lead".lead_type = ?`, projectID, "Deal")

	if tenantID != nil {
		query = query.Where(`"Lead".tenant_id = ?`, *tenantID)
	}

	err := query.Find(&deals).Error
	return deals, err
}
