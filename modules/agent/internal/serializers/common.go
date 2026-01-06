package serializers

// PagedResponse represents a paginated API response
type PagedResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"totalCount"`
	Page       int         `json:"currentPage"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// NewPagedResponse creates a paged response
func NewPagedResponse(items interface{}, totalCount int64, page, pageSize, totalPages int) PagedResponse {
	return PagedResponse{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a simple success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
