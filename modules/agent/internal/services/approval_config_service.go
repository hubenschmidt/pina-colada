package services

import (
	"errors"

	"agent/internal/repositories"
	"agent/internal/serializers"
)

var ErrInvalidEntityTypeConfig = errors.New("invalid entity type")

// ApprovalConfigService handles approval config business logic
type ApprovalConfigService struct {
	configRepo *repositories.ApprovalConfigRepository
}

// NewApprovalConfigService creates a new approval config service
func NewApprovalConfigService(configRepo *repositories.ApprovalConfigRepository) *ApprovalConfigService {
	return &ApprovalConfigService{configRepo: configRepo}
}

// GetConfig returns approval config for all entities for a tenant
func (s *ApprovalConfigService) GetConfig(tenantID int64) ([]serializers.ApprovalConfigResponse, error) {
	configs, err := s.configRepo.FindAllByTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// Build map of existing configs
	configMap := make(map[string]bool)
	for _, c := range configs {
		configMap[c.EntityType] = c.RequiresApproval
	}

	// Return all entities with their approval status (default true if not configured)
	result := make([]serializers.ApprovalConfigResponse, len(repositories.SupportedEntities))
	for i, entityType := range repositories.SupportedEntities {
		requiresApproval := true
		if val, exists := configMap[entityType]; exists {
			requiresApproval = val
		}
		result[i] = serializers.ApprovalConfigResponse{
			EntityType:       entityType,
			RequiresApproval: requiresApproval,
		}
	}

	return result, nil
}

// UpdateConfig updates approval config for an entity type
func (s *ApprovalConfigService) UpdateConfig(tenantID int64, entityType string, requiresApproval bool) (*serializers.ApprovalConfigResponse, error) {
	if !isValidEntityTypeForConfig(entityType) {
		return nil, ErrInvalidEntityTypeConfig
	}

	if err := s.configRepo.Upsert(tenantID, entityType, requiresApproval); err != nil {
		return nil, err
	}

	return &serializers.ApprovalConfigResponse{
		EntityType:       entityType,
		RequiresApproval: requiresApproval,
	}, nil
}

// RequiresApproval checks if an entity type requires approval
func (s *ApprovalConfigService) RequiresApproval(tenantID int64, entityType string) (bool, error) {
	return s.configRepo.RequiresApproval(tenantID, entityType)
}

func isValidEntityTypeForConfig(entityType string) bool {
	for _, e := range repositories.SupportedEntities {
		if e == entityType {
			return true
		}
	}
	return false
}
