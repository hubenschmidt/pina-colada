package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// AvailableModelDTO for service layer (services don't import models)
type AvailableModelDTO struct {
	ID                    int64
	Provider              string
	ModelName             string
	DisplayName           string
	SortOrder             int
	IsActive              bool
	DefaultTimeoutSeconds int
}

type AvailableModelRepository struct {
	db *gorm.DB
}

func NewAvailableModelRepository(db *gorm.DB) *AvailableModelRepository {
	return &AvailableModelRepository{db: db}
}

func (r *AvailableModelRepository) GetActiveModels() ([]AvailableModelDTO, error) {
	var modelList []models.AvailableModel
	err := r.db.Where("is_active = ?", true).Order("provider, sort_order").Find(&modelList).Error
	if err != nil {
		return nil, err
	}
	return toAvailableModelDTOs(modelList), nil
}

func (r *AvailableModelRepository) GetModelsByProvider(provider string) ([]AvailableModelDTO, error) {
	var modelList []models.AvailableModel
	err := r.db.Where("provider = ? AND is_active = ?", provider, true).Order("sort_order").Find(&modelList).Error
	if err != nil {
		return nil, err
	}
	return toAvailableModelDTOs(modelList), nil
}

func (r *AvailableModelRepository) GetModel(provider, modelName string) (*AvailableModelDTO, error) {
	var model models.AvailableModel
	err := r.db.Where("provider = ? AND model_name = ?", provider, modelName).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // Not found = nil, nil per CSR rules
	}
	if err != nil {
		return nil, err
	}
	return toAvailableModelDTO(&model), nil
}

func (r *AvailableModelRepository) GetModelByName(modelName string) (*AvailableModelDTO, error) {
	var model models.AvailableModel
	err := r.db.Where("model_name = ? AND is_active = ?", modelName, true).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toAvailableModelDTO(&model), nil
}

func toAvailableModelDTO(m *models.AvailableModel) *AvailableModelDTO {
	return &AvailableModelDTO{
		ID:                    m.ID,
		Provider:              m.Provider,
		ModelName:             m.ModelName,
		DisplayName:           m.DisplayName,
		SortOrder:             m.SortOrder,
		IsActive:              m.IsActive,
		DefaultTimeoutSeconds: m.DefaultTimeoutSeconds,
	}
}

func toAvailableModelDTOs(modelList []models.AvailableModel) []AvailableModelDTO {
	dtos := make([]AvailableModelDTO, len(modelList))
	for i, m := range modelList {
		dtos[i] = *toAvailableModelDTO(&m)
	}
	return dtos
}
