package serializers

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
