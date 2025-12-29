package serializers

import (
	"time"

	"agent/internal/models"
)

// TechnologyResponse represents a technology
type TechnologyResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Vendor    *string   `json:"vendor"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TechnologyToResponse converts a model to response
func TechnologyToResponse(m *models.Technology) TechnologyResponse {
	return TechnologyResponse{
		ID:        m.ID,
		Name:      m.Name,
		Category:  m.Category,
		Vendor:    m.Vendor,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// TechnologiesToResponse converts a slice of models to responses
func TechnologiesToResponse(models []models.Technology) []TechnologyResponse {
	resp := make([]TechnologyResponse, len(models))
	for i := range models {
		resp[i] = TechnologyToResponse(&models[i])
	}
	return resp
}
