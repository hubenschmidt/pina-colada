package schemas

// ProvenanceCreate represents the request body for creating a provenance record
type ProvenanceCreate struct {
	EntityType string   `json:"entity_type" validate:"required"`
	EntityID   int64    `json:"entity_id" validate:"required"`
	FieldName  string   `json:"field_name" validate:"required"`
	Source     string   `json:"source" validate:"required"`
	SourceURL  *string  `json:"source_url" validate:"omitempty,url"`
	Confidence *float64 `json:"confidence" validate:"omitempty,min=0,max=1"`
	RawValue   any      `json:"raw_value"`
}
