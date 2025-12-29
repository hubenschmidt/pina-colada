package serializers

import (
	"time"

	"agent/internal/models"
)

// NoteResponse represents a note
type NoteResponse struct {
	ID         int64     `json:"id"`
	TenantID   int64     `json:"tenant_id"`
	EntityType string    `json:"entity_type"`
	EntityID   int64     `json:"entity_id"`
	Content    string    `json:"content"`
	CreatedBy  int64     `json:"created_by"`
	UpdatedBy  int64     `json:"updated_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NoteToResponse converts a model to response
func NoteToResponse(m *models.Note) NoteResponse {
	return NoteResponse{
		ID:         m.ID,
		TenantID:   m.TenantID,
		EntityType: m.EntityType,
		EntityID:   m.EntityID,
		Content:    m.Content,
		CreatedBy:  m.CreatedBy,
		UpdatedBy:  m.UpdatedBy,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// NotesToResponse converts a slice of models to responses
func NotesToResponse(models []models.Note) []NoteResponse {
	resp := make([]NoteResponse, len(models))
	for i := range models {
		resp[i] = NoteToResponse(&models[i])
	}
	return resp
}
