package schemas

// NoteCreate represents the request body for creating a note
type NoteCreate struct {
	EntityType string `json:"entity_type" validate:"required"`
	EntityID   int64  `json:"entity_id" validate:"required"`
	Content    string `json:"content" validate:"required"`
}

// NoteUpdate represents the request body for updating a note
type NoteUpdate struct {
	Content string `json:"content" validate:"required"`
}
