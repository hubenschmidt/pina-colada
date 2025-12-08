package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// IndividualRepository handles individual data access
type IndividualRepository struct {
	db *gorm.DB
}

// NewIndividualRepository creates a new individual repository
func NewIndividualRepository(db *gorm.DB) *IndividualRepository {
	return &IndividualRepository{db: db}
}

// FindAll returns paginated individuals
func (r *IndividualRepository) FindAll(params PaginationParams, tenantID *int64) (*PaginatedResult[models.Individual], error) {
	var individuals []models.Individual
	var totalCount int64

	query := r.db.Model(&models.Individual{}).
		Preload("Account").
		Preload("Account.Industries").
		Preload("Industries").
		Preload("Contacts")

	if tenantID != nil {
		query = query.Where("individuals.tenant_id = ?", *tenantID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(
			"LOWER(individuals.first_name) LIKE LOWER(?) OR LOWER(individuals.last_name) LIKE LOWER(?) OR LOWER(individuals.email) LIKE LOWER(?)",
			searchTerm, searchTerm, searchTerm,
		)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&individuals).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[models.Individual]{
		Items:      individuals,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindByID returns an individual by ID
func (r *IndividualRepository) FindByID(id int64) (*models.Individual, error) {
	var ind models.Individual
	err := r.db.
		Preload("Account").
		Preload("Account.Industries").
		Preload("Industries").
		Preload("Contacts").
		Preload("Projects").
		First(&ind, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &ind, nil
}

// Create creates a new individual
func (r *IndividualRepository) Create(ind *models.Individual) error {
	return r.db.Create(ind).Error
}

// Update updates an individual
func (r *IndividualRepository) Update(ind *models.Individual, updates map[string]interface{}) error {
	return r.db.Model(ind).Updates(updates).Error
}

// Delete deletes an individual by ID
func (r *IndividualRepository) Delete(id int64) error {
	return r.db.Delete(&models.Individual{}, id).Error
}

// Search searches individuals by name or email
func (r *IndividualRepository) Search(query string, tenantID *int64, limit int) ([]models.Individual, error) {
	var individuals []models.Individual
	searchTerm := "%" + query + "%"

	q := r.db.Model(&models.Individual{}).
		Preload("Account").
		Where(
			"LOWER(first_name) LIKE LOWER(?) OR LOWER(last_name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)",
			searchTerm, searchTerm, searchTerm,
		)

	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}

	err := q.Limit(limit).Find(&individuals).Error
	return individuals, err
}
