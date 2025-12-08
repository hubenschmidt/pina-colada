package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

// IndividualService handles individual business logic
type IndividualService struct {
	indRepo *repositories.IndividualRepository
}

// NewIndividualService creates a new individual service
func NewIndividualService(indRepo *repositories.IndividualRepository) *IndividualService {
	return &IndividualService{indRepo: indRepo}
}

// GetIndividuals returns paginated individuals
func (s *IndividualService) GetIndividuals(page, pageSize int, orderBy, order, search string, tenantID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	result, err := s.indRepo.FindAll(params, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.IndividualListResponse, len(result.Items))
	for i := range result.Items {
		items[i] = serializers.IndividualToListResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// GetIndividual returns an individual by ID
func (s *IndividualService) GetIndividual(id int64) (*serializers.IndividualDetailResponse, error) {
	ind, err := s.indRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if ind == nil {
		return nil, errors.New("individual not found")
	}

	resp := serializers.IndividualToDetailResponse(ind)
	return &resp, nil
}

// SearchIndividuals searches individuals by name or email
func (s *IndividualService) SearchIndividuals(query string, tenantID *int64, limit int) ([]serializers.IndividualBrief, error) {
	if limit <= 0 {
		limit = 10
	}

	individuals, err := s.indRepo.Search(query, tenantID, limit)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.IndividualBrief, len(individuals))
	for i := range individuals {
		results[i] = serializers.IndividualBrief{
			ID:        individuals[i].ID,
			FirstName: individuals[i].FirstName,
			LastName:  individuals[i].LastName,
			Email:     individuals[i].Email,
		}
	}

	return results, nil
}

// DeleteIndividual deletes an individual
func (s *IndividualService) DeleteIndividual(id int64) error {
	ind, err := s.indRepo.FindByID(id)
	if err != nil {
		return err
	}
	if ind == nil {
		return errors.New("individual not found")
	}

	return s.indRepo.Delete(id)
}
