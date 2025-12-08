package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

// OrganizationService handles organization business logic
type OrganizationService struct {
	orgRepo *repositories.OrganizationRepository
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(orgRepo *repositories.OrganizationRepository) *OrganizationService {
	return &OrganizationService{orgRepo: orgRepo}
}

// GetOrganizations returns paginated organizations
func (s *OrganizationService) GetOrganizations(page, pageSize int, orderBy, order, search string, tenantID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	result, err := s.orgRepo.FindAll(params, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.OrganizationListResponse, len(result.Items))
	for i := range result.Items {
		items[i] = serializers.OrganizationToListResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// GetOrganization returns an organization by ID
func (s *OrganizationService) GetOrganization(id int64) (*serializers.OrganizationDetailResponse, error) {
	org, err := s.orgRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("organization not found")
	}

	resp := serializers.OrganizationToDetailResponse(org)
	return &resp, nil
}

// SearchOrganizations searches organizations by name
func (s *OrganizationService) SearchOrganizations(query string, tenantID *int64, limit int) ([]serializers.OrganizationBrief, error) {
	if limit <= 0 {
		limit = 10
	}

	orgs, err := s.orgRepo.FindByName(query, tenantID, limit)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.OrganizationBrief, len(orgs))
	for i := range orgs {
		results[i] = serializers.OrganizationBrief{
			ID:   orgs[i].ID,
			Name: orgs[i].Name,
		}
	}

	return results, nil
}

// DeleteOrganization deletes an organization
func (s *OrganizationService) DeleteOrganization(id int64) error {
	org, err := s.orgRepo.FindByID(id)
	if err != nil {
		return err
	}
	if org == nil {
		return errors.New("organization not found")
	}

	return s.orgRepo.Delete(id)
}
