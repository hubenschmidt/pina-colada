package schemas

// DocumentUpdate represents the request body for updating a document
type DocumentUpdate struct {
	Description *string `json:"description"`
}

// EntityLink represents a link between a document and an entity
type EntityLink struct {
	EntityType string `json:"entity_type" validate:"required"`
	EntityID   int64  `json:"entity_id" validate:"required"`
}

// DocumentResponse represents a document in API responses
type DocumentResponse struct {
	ID          int64   `json:"id"`
	TenantID    int64   `json:"tenant_id"`
	Filename    string  `json:"filename"`
	ContentType string  `json:"content_type"`
	Description *string `json:"description"`
	FileSize    int64   `json:"file_size"`
	Summary     []byte  `json:"summary,omitempty"`
}
