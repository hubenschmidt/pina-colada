package repositories

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// GenericEntityRepository provides generic entity operations
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
		Order("id DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// searchableColumns are columns we attempt to search (if they exist)
var searchableColumns = []string{"name", "title", "email", "first_name", "last_name", "description", "job_title"}

// SearchEntities searches entities by text fields that exist on the table
func (r *GenericEntityRepository) SearchEntities(entityType, query string, limit int) ([]map[string]interface{}, error) {
	tableName := toTableName(entityType)

	if query == "" {
		return r.ListEntities(entityType, limit)
	}

	// Get actual columns for this table
	cols, err := r.getTableColumns(tableName)
	if err != nil {
		return nil, err
	}

	// Build OR conditions for columns that exist
	searchPattern := "%" + query + "%"
	var conditions []string
	var args []interface{}
	for _, col := range searchableColumns {
		if cols[col] {
			conditions = append(conditions, fmt.Sprintf(`"%s" ILIKE ?`, col))
			args = append(args, searchPattern)
		}
	}

	if len(conditions) == 0 {
		return r.ListEntities(entityType, limit)
	}

	var results []map[string]interface{}
	err = r.db.Table(fmt.Sprintf(`"%s"`, tableName)).
		Where(strings.Join(conditions, " OR "), args...).
		Limit(limit).
		Order("id DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// getTableColumns returns a set of searchable column names that exist on the table
func (r *GenericEntityRepository) getTableColumns(tableName string) (map[string]bool, error) {
	migrator := r.db.Migrator()
	result := make(map[string]bool)
	for _, col := range searchableColumns {
		if migrator.HasColumn(&tableRef{name: tableName}, col) {
			result[col] = true
		}
	}
	return result, nil
}

// tableRef implements tabler interface for dynamic table names
type tableRef struct {
	name string
}

func (t *tableRef) TableName() string {
	return t.name
}

// toTableName converts snake_case entity type to PascalCase table name
func toTableName(entityType string) string {
	parts := strings.Split(entityType, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "_")
}
