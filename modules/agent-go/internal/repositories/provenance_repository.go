package repositories

import (
	"encoding/json"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProvenanceRepository handles data provenance data access
type ProvenanceRepository struct {
	db *gorm.DB
}

// NewProvenanceRepository creates a new provenance repository
func NewProvenanceRepository(db *gorm.DB) *ProvenanceRepository {
	return &ProvenanceRepository{db: db}
}

// FindByEntity returns provenance records for an entity
func (r *ProvenanceRepository) FindByEntity(entityType string, entityID int64, fieldName *string) ([]models.DataProvenance, error) {
	var records []models.DataProvenance

	query := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID)
	if fieldName != nil && *fieldName != "" {
		query = query.Where("field_name = ?", *fieldName)
	}

	err := query.Order("verified_at DESC").Find(&records).Error
	return records, err
}

// ProvenanceCreateInput contains data needed to create a provenance record
type ProvenanceCreateInput struct {
	EntityType string
	EntityID   int64
	FieldName  string
	Source     string
	SourceURL  *string
	Confidence *float64
	VerifiedBy *int64
	RawValue   any
}

// Create creates a new provenance record
func (r *ProvenanceRepository) Create(input ProvenanceCreateInput) (*models.DataProvenance, error) {
	provenance := &models.DataProvenance{
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		FieldName:  input.FieldName,
		Source:     input.Source,
		SourceURL:  input.SourceURL,
		VerifiedBy: input.VerifiedBy,
	}

	if input.Confidence != nil {
		conf := decimal.NewFromFloat(*input.Confidence)
		provenance.Confidence = &conf
	}

	rawJSON, err := marshalRawValue(input.RawValue)
	if err != nil {
		return nil, err
	}
	provenance.RawValue = rawJSON

	if err := r.db.Create(provenance).Error; err != nil {
		return nil, err
	}

	return provenance, nil
}

func marshalRawValue(rawValue interface{}) (datatypes.JSON, error) {
	if rawValue == nil {
		return nil, nil
	}
	rawBytes, err := json.Marshal(rawValue)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(rawBytes), nil
}
