package repositories

import (
	"errors"

	apperrors "agent/internal/errors"
	"agent/internal/models"
	"gorm.io/gorm"
)

// TechnologyRepository handles technology data access
type TechnologyRepository struct {
	db *gorm.DB
}

// NewTechnologyRepository creates a new technology repository
func NewTechnologyRepository(db *gorm.DB) *TechnologyRepository {
	return &TechnologyRepository{db: db}
}

// FindAll returns all technologies, optionally filtered by category
func (r *TechnologyRepository) FindAll(category *string) ([]models.Technology, error) {
	var techs []models.Technology

	query := r.db.Order("name ASC")
	if category != nil && *category != "" {
		query = query.Where("category = ?", *category)
	}

	err := query.Find(&techs).Error
	return techs, err
}

// FindByID returns a technology by ID
func (r *TechnologyRepository) FindByID(id int64) (*models.Technology, error) {
	var tech models.Technology
	err := r.db.First(&tech, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tech, nil
}

// TechnologyCreateInput contains data needed to create a technology
type TechnologyCreateInput struct {
	Name     string
	Category string
	Vendor   *string
}

// Create creates a new technology
func (r *TechnologyRepository) Create(input TechnologyCreateInput) (*models.Technology, error) {
	tech := &models.Technology{
		Name:     input.Name,
		Category: input.Category,
		Vendor:   input.Vendor,
	}

	if err := r.db.Create(tech).Error; err != nil {
		return nil, err
	}

	return tech, nil
}
