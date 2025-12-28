package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// ProvenanceResponse represents a data provenance record
type ProvenanceResponse struct {
	ID         int64            `json:"id"`
	EntityType string           `json:"entity_type"`
	EntityID   int64            `json:"entity_id"`
	FieldName  string           `json:"field_name"`
	Source     string           `json:"source"`
	SourceURL  *string          `json:"source_url"`
	Confidence *decimal.Decimal `json:"confidence"`
	VerifiedAt time.Time        `json:"verified_at"`
	VerifiedBy *int64           `json:"verified_by"`
	RawValue   datatypes.JSON   `json:"raw_value"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// ProvenanceToResponse converts a model to response
func ProvenanceToResponse(m *models.DataProvenance) ProvenanceResponse {
	return ProvenanceResponse{
		ID:         m.ID,
		EntityType: m.EntityType,
		EntityID:   m.EntityID,
		FieldName:  m.FieldName,
		Source:     m.Source,
		SourceURL:  m.SourceURL,
		Confidence: m.Confidence,
		VerifiedAt: m.VerifiedAt,
		VerifiedBy: m.VerifiedBy,
		RawValue:   m.RawValue,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// ProvenancesToResponse converts a slice of models to responses
func ProvenancesToResponse(models []models.DataProvenance) []ProvenanceResponse {
	resp := make([]ProvenanceResponse, len(models))
	for i := range models {
		resp[i] = ProvenanceToResponse(&models[i])
	}
	return resp
}
