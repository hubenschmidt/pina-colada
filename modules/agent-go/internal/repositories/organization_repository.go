package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// OrganizationRepository handles organization data access
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// FindAll returns paginated organizations
func (r *OrganizationRepository) FindAll(params PaginationParams, tenantID *int64) (*PaginatedResult[models.Organization], error) {
	var orgs []models.Organization
	var totalCount int64

	query := r.db.Model(&models.Organization{}).
		Preload("Account").
		Preload("Account.Industries").
		Preload("Technologies")

	// Filter by tenant
	if tenantID != nil {
		query = query.Where("organizations.tenant_id = ?", *tenantID)
	}

	// Apply search
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("LOWER(organizations.name) LIKE LOWER(?)", searchTerm)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	query = ApplyPagination(query, params)

	if err := query.Find(&orgs).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[models.Organization]{
		Items:      orgs,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindByID returns an organization by ID
func (r *OrganizationRepository) FindByID(id int64) (*models.Organization, error) {
	var org models.Organization
	err := r.db.
		Preload("Account").
		Preload("Account.Industries").
		Preload("Technologies").
		Preload("FundingRounds").
		First(&org, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// FindByName searches organizations by name
func (r *OrganizationRepository) FindByName(name string, tenantID *int64, limit int) ([]models.Organization, error) {
	var orgs []models.Organization

	query := r.db.Model(&models.Organization{}).
		Preload("Account").
		Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%")

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	err := query.Limit(limit).Find(&orgs).Error
	return orgs, err
}

// Create creates a new organization
func (r *OrganizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}

// Update updates an organization
func (r *OrganizationRepository) Update(org *models.Organization, updates map[string]interface{}) error {
	return r.db.Model(org).Updates(updates).Error
}

// Delete deletes an organization by ID
func (r *OrganizationRepository) Delete(id int64) error {
	return r.db.Delete(&models.Organization{}, id).Error
}
