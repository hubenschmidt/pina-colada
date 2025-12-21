package repositories

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// GenericEntityRepository provides generic entity listing
type GenericEntityRepository struct {
	db *gorm.DB
}

// NewGenericEntityRepository creates a new generic entity repository
func NewGenericEntityRepository(db *gorm.DB) *GenericEntityRepository {
	return &GenericEntityRepository{db: db}
}

// ListEntities lists entities from any table by entity type name
func (r *GenericEntityRepository) ListEntities(entityType string, limit int) ([]map[string]interface{}, error) {
	tableName := toTableName(entityType)

	var results []map[string]interface{}
	err := r.db.Table(fmt.Sprintf(`"%s"`, tableName)).
		Limit(limit).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// toTableName converts snake_case entity type to PascalCase table name
// e.g., "funding_round" -> "Funding_Round", "deal" -> "Deal"
func toTableName(entityType string) string {
	parts := strings.Split(entityType, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "_")
}
