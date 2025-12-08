package schemas

// MarkReadRequest represents the request body for marking notifications as read
type MarkReadRequest struct {
	NotificationIDs []int64 `json:"notification_ids" validate:"required,min=1"`
}

// MarkEntityReadRequest represents the request body for marking entity notifications as read
type MarkEntityReadRequest struct {
	EntityType string `json:"entity_type" validate:"required"`
	EntityID   int64  `json:"entity_id" validate:"required"`
}
