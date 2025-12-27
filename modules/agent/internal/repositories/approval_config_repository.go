package repositories

import (
	"errors"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// ApprovalConfigRepository handles entity approval config data access
type ApprovalConfigRepository struct {
	db *gorm.DB
}

// NewApprovalConfigRepository creates a new approval config repository
func NewApprovalConfigRepository(db *gorm.DB) *ApprovalConfigRepository {
	return &ApprovalConfigRepository{db: db}
}

// FindByTenantAndEntity returns config for a specific tenant and entity type
func (r *ApprovalConfigRepository) FindByTenantAndEntity(tenantID int64, entityType string) (*models.EntityApprovalConfig, error) {
	var config models.EntityApprovalConfig
	err := r.db.Where("tenant_id = ? AND entity_type = ?", tenantID, entityType).First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindAllByTenant returns all approval configs for a tenant
func (r *ApprovalConfigRepository) FindAllByTenant(tenantID int64) ([]models.EntityApprovalConfig, error) {
	var configs []models.EntityApprovalConfig
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Upsert creates or updates an approval config
func (r *ApprovalConfigRepository) Upsert(tenantID int64, entityType string, requiresApproval bool) error {
	config := models.EntityApprovalConfig{
		TenantID:         tenantID,
		EntityType:       entityType,
		RequiresApproval: requiresApproval,
	}

	return r.db.Where("tenant_id = ? AND entity_type = ?", tenantID, entityType).
		Assign(models.EntityApprovalConfig{RequiresApproval: requiresApproval}).
		FirstOrCreate(&config).Error
}

// RequiresApproval checks if an entity type requires approval for a tenant
// Returns true if no config exists (default behavior)
func (r *ApprovalConfigRepository) RequiresApproval(tenantID int64, entityType string) (bool, error) {
	var config models.EntityApprovalConfig
	err := r.db.Where("tenant_id = ? AND entity_type = ?", tenantID, entityType).First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil // Default: requires approval
	}
	if err != nil {
		return false, err
	}
	return config.RequiresApproval, nil
}
