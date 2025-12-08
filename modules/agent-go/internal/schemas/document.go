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
