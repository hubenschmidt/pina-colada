package schemas

// CommentCreate represents the request body for creating a comment
type CommentCreate struct {
	CommentableType string `json:"commentable_type" validate:"required"`
	CommentableID   int64  `json:"commentable_id" validate:"required"`
	Content         string `json:"content" validate:"required"`
	ParentCommentID *int64 `json:"parent_comment_id"`
}

// CommentUpdate represents the request body for updating a comment
type CommentUpdate struct {
	Content string `json:"content" validate:"required"`
}
