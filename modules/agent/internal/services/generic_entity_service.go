package services

import (
	"agent/internal/repositories"
)

// GenericEntityService provides generic entity operations
type GenericEntityService struct {
	repo *repositories.GenericEntityRepository
}

// NewGenericEntityService creates a new generic entity service
func NewGenericEntityService(repo *repositories.GenericEntityRepository) *GenericEntityService {
	return &GenericEntityService{repo: repo}
}

// ListEntities lists entities from any table by entity type name
func (s *GenericEntityService) ListEntities(entityType string, limit int) ([]map[string]interface{}, error) {
	return s.repo.ListEntities(entityType, limit)
}

// SearchEntities searches entities by query string
func (s *GenericEntityService) SearchEntities(entityType, query string, limit int) ([]map[string]interface{}, error) {
	return s.repo.SearchEntities(entityType, query, limit)
}
