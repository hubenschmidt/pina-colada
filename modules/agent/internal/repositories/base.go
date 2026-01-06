package repositories

import (
	"fmt"

	"gorm.io/gorm"
)

// PaginationParams contains common pagination parameters
type PaginationParams struct {
	Page        int
	PageSize    int
	OrderBy     string
	Order       string // ASC or DESC
	Search      string
	CurrentOnly bool
}

// PaginatedResult contains paginated query results
type PaginatedResult[T any] struct {
	Items      []T
	TotalCount int64
	Page       int
	PageSize   int
	TotalPages int
}

// NewPaginationParams creates pagination params with defaults
func NewPaginationParams(page, pageSize int, orderBy, order string) PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 500 {
		pageSize = 500
	}
	if order == "" {
		order = "DESC"
	}
	if orderBy == "" {
		orderBy = "updated_at"
	}
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		OrderBy:  orderBy,
		Order:    order,
	}
}

// ApplyPagination applies pagination to a GORM query
func ApplyPagination(db *gorm.DB, params PaginationParams) *gorm.DB {
	offset := (params.Page - 1) * params.PageSize

	orderClause := fmt.Sprintf("%s %s", params.OrderBy, params.Order)

	return db.Order(orderClause).Offset(offset).Limit(params.PageSize)
}

// CountTotal counts total records matching the query
func CountTotal(db *gorm.DB, model interface{}) (int64, error) {
	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CalculateTotalPages calculates total pages from count and page size
func CalculateTotalPages(totalCount int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		pages++
	}
	return pages
}
