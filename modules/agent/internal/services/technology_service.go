package services

import (
	"errors"

	"agent/internal/repositories"
	"agent/internal/schemas"
	"agent/internal/serializers"
)

var ErrTechnologyNotFound = errors.New("technology not found")

// TechnologyService handles technology business logic
type TechnologyService struct {
	techRepo *repositories.TechnologyRepository
}

// NewTechnologyService creates a new technology service
func NewTechnologyService(techRepo *repositories.TechnologyRepository) *TechnologyService {
	return &TechnologyService{techRepo: techRepo}
}

// GetAllTechnologies returns all technologies, optionally filtered by category
func (s *TechnologyService) GetAllTechnologies(category *string) ([]serializers.TechnologyResponse, error) {
	techs, err := s.techRepo.FindAll(category)
	if err != nil {
		return nil, err
	}
	return serializers.TechnologiesToResponse(techs), nil
}

// GetTechnology returns a technology by ID
func (s *TechnologyService) GetTechnology(id int64) (*serializers.TechnologyResponse, error) {
	tech, err := s.techRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if tech == nil {
		return nil, ErrTechnologyNotFound
	}
	resp := serializers.TechnologyToResponse(tech)
	return &resp, nil
}

// CreateTechnology creates a new technology
func (s *TechnologyService) CreateTechnology(input schemas.TechnologyCreate) (*serializers.TechnologyResponse, error) {
	repoInput := repositories.TechnologyCreateInput{
		Name:     input.Name,
		Category: input.Category,
		Vendor:   input.Vendor,
	}
	tech, err := s.techRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}
	resp := serializers.TechnologyToResponse(tech)
	return &resp, nil
}
