package services

import (
	"errors"

	"agent/internal/repositories"
	"agent/internal/schemas"
	"agent/internal/serializers"
)

var ErrInvalidEntityType = errors.New("invalid entity_type: must be Organization or Individual")

// ProvenanceService handles provenance business logic
type ProvenanceService struct {
	provenanceRepo *repositories.ProvenanceRepository
}

// NewProvenanceService creates a new provenance service
func NewProvenanceService(provenanceRepo *repositories.ProvenanceRepository) *ProvenanceService {
	return &ProvenanceService{provenanceRepo: provenanceRepo}
}

func (s *ProvenanceService) isValidEntityType(entityType string) bool {
	return entityType == "Organization" || entityType == "Individual"
}

// GetProvenance returns provenance records for an entity
func (s *ProvenanceService) GetProvenance(entityType string, entityID int64, fieldName *string) ([]serializers.ProvenanceResponse, error) {
	if !s.isValidEntityType(entityType) {
		return nil, ErrInvalidEntityType
	}
	records, err := s.provenanceRepo.FindByEntity(entityType, entityID, fieldName)
	if err != nil {
		return nil, err
	}
	return serializers.ProvenancesToResponse(records), nil
}

// CreateProvenance creates a new provenance record
func (s *ProvenanceService) CreateProvenance(input schemas.ProvenanceCreate, userID *int64) (*serializers.ProvenanceResponse, error) {
	if !s.isValidEntityType(input.EntityType) {
		return nil, ErrInvalidEntityType
	}

	repoInput := repositories.ProvenanceCreateInput{
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		FieldName:  input.FieldName,
		Source:     input.Source,
		SourceURL:  input.SourceURL,
		Confidence: input.Confidence,
		VerifiedBy: userID,
		RawValue:   input.RawValue,
	}

	record, err := s.provenanceRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}
	resp := serializers.ProvenanceToResponse(record)
	return &resp, nil
}
