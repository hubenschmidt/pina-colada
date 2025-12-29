package serializers

import (
	"strings"
	"time"

	"agent/internal/models"
)

// CommentResponse represents a comment response
type CommentResponse struct {
	ID              int64     `json:"id"`
	TenantID        int64     `json:"tenant_id"`
	CommentableType string    `json:"commentable_type"`
	CommentableID   int64     `json:"commentable_id"`
	Content         string    `json:"content"`
	CreatedBy       int64     `json:"created_by"`
	CreatedByName   *string   `json:"created_by_name"`
	CreatedByEmail  *string   `json:"created_by_email"`
	IndividualID    *int64    `json:"individual_id"`
	ParentCommentID *int64    `json:"parent_comment_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CommentToResponse converts Comment model to response
func CommentToResponse(comment *models.Comment) CommentResponse {
	resp := CommentResponse{
		ID:              comment.ID,
		TenantID:        comment.TenantID,
		CommentableType: comment.CommentableType,
		CommentableID:   comment.CommentableID,
		Content:         comment.Content,
		CreatedBy:       comment.CreatedBy,
		ParentCommentID: comment.ParentCommentID,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
	}

	resp.CreatedByName, resp.CreatedByEmail, resp.IndividualID = extractCreatorInfo(comment.Creator)

	return resp
}

func extractCreatorInfo(creator *models.User) (*string, *string, *int64) {
	if creator == nil {
		return nil, nil, nil
	}

	name := buildUserFullName(creator.FirstName, creator.LastName)
	return name, &creator.Email, creator.IndividualID
}

func buildUserFullName(firstName, lastName *string) *string {
	var parts []string

	if firstName != nil && *firstName != "" {
		parts = append(parts, *firstName)
	}
	if lastName != nil && *lastName != "" {
		parts = append(parts, *lastName)
	}

	if len(parts) == 0 {
		return nil
	}

	name := strings.Join(parts, " ")
	return &name
}

// CommentsToResponse converts a slice of Comment models to responses
func CommentsToResponse(comments []models.Comment) []CommentResponse {
	result := make([]CommentResponse, len(comments))
	for i := range comments {
		result[i] = CommentToResponse(&comments[i])
	}
	return result
}
