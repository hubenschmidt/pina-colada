package repositories

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

type LookupRepository struct {
	db *gorm.DB
}

func NewLookupRepository(db *gorm.DB) *LookupRepository {
	return &LookupRepository{db: db}
}

func (r *LookupRepository) FindAllIndustries() ([]models.Industry, error) {
	var industries []models.Industry
	err := r.db.Order("name ASC").Find(&industries).Error
	return industries, err
}

func (r *LookupRepository) FindAllEmployeeCountRanges() ([]models.EmployeeCountRange, error) {
	var ranges []models.EmployeeCountRange
	err := r.db.Order("display_order ASC").Find(&ranges).Error
	return ranges, err
}

func (r *LookupRepository) FindAllRevenueRanges() ([]models.RevenueRange, error) {
	var ranges []models.RevenueRange
	err := r.db.Order("display_order ASC").Find(&ranges).Error
	return ranges, err
}

func (r *LookupRepository) FindAllFundingStages() ([]models.FundingStage, error) {
	var stages []models.FundingStage
	err := r.db.Order("display_order ASC").Find(&stages).Error
	return stages, err
}

func (r *LookupRepository) FindTaskStatuses() ([]models.Status, error) {
	var statuses []models.Status
	err := r.db.Where("category = ?", "task_status").Order("name ASC").Find(&statuses).Error
	return statuses, err
}

func (r *LookupRepository) FindTaskPriorities() ([]models.Status, error) {
	var priorities []models.Status
	err := r.db.Where("category = ?", "task_priority").Order("name ASC").Find(&priorities).Error
	return priorities, err
}

func (r *LookupRepository) FindAllSalaryRanges() ([]models.SalaryRange, error) {
	var ranges []models.SalaryRange
	err := r.db.Order("display_order ASC").Find(&ranges).Error
	return ranges, err
}

// FindStatusByName finds a status by name and category
func (r *LookupRepository) FindStatusByName(name, category string) (*models.Status, error) {
	var status models.Status
	err := r.db.Where("LOWER(name) = LOWER(?) AND category = ?", name, category).First(&status).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// CreateIndustry creates a new industry
func (r *LookupRepository) CreateIndustry(name string) (*models.Industry, error) {
	industry := &models.Industry{Name: name}
	if err := r.db.Create(industry).Error; err != nil {
		return nil, err
	}
	return industry, nil
}
