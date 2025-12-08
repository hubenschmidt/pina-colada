package serializers

import (
	"strings"
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
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

	if comment.Creator != nil {
		var nameParts []string
		if comment.Creator.FirstName != nil && *comment.Creator.FirstName != "" {
			nameParts = append(nameParts, *comment.Creator.FirstName)
		}
		if comment.Creator.LastName != nil && *comment.Creator.LastName != "" {
			nameParts = append(nameParts, *comment.Creator.LastName)
		}
		if len(nameParts) > 0 {
			name := strings.Join(nameParts, " ")
			resp.CreatedByName = &name
		}
		resp.CreatedByEmail = &comment.Creator.Email
		resp.IndividualID = comment.Creator.IndividualID
	}

	return resp
}

// CommentsToResponse converts a slice of Comment models to responses
func CommentsToResponse(comments []models.Comment) []CommentResponse {
	result := make([]CommentResponse, len(comments))
	for i := range comments {
		result[i] = CommentToResponse(&comments[i])
	}
	return result
}
