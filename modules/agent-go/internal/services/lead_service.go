package services

import (
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

// LeadService handles lead business logic (opportunities, partnerships)
type LeadService struct {
	leadRepo *repositories.LeadRepository
}

// NewLeadService creates a new lead service
func NewLeadService(leadRepo *repositories.LeadRepository) *LeadService {
	return &LeadService{leadRepo: leadRepo}
}

// GetOpportunities returns paginated opportunities
func (s *LeadService) GetOpportunities(page, pageSize int, orderBy, order, search string, tenantID, projectID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	result, err := s.leadRepo.FindAllOpportunities(params, tenantID, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.OpportunityResponse, len(result.Items))
	for i := range result.Items {
		items[i] = serializers.OpportunityToResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// GetOpportunity returns a single opportunity by ID
func (s *LeadService) GetOpportunity(id int64) (*serializers.OpportunityResponse, error) {
	opp, err := s.leadRepo.FindOpportunityByID(id)
	if err != nil {
		return nil, err
	}
	if opp == nil {
		return nil, nil
	}

	resp := serializers.OpportunityToResponse(opp)
	return &resp, nil
}

// GetPartnerships returns paginated partnerships
func (s *LeadService) GetPartnerships(page, pageSize int, orderBy, order, search string, tenantID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	result, err := s.leadRepo.FindAllPartnerships(params, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.PartnershipResponse, len(result.Items))
	for i := range result.Items {
		items[i] = serializers.PartnershipToResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// GetPartnership returns a single partnership by ID
func (s *LeadService) GetPartnership(id int64) (*serializers.PartnershipResponse, error) {
	partnership, err := s.leadRepo.FindPartnershipByID(id)
	if err != nil {
		return nil, err
	}
	if partnership == nil {
		return nil, nil
	}

	resp := serializers.PartnershipToResponse(partnership)
	return &resp, nil
}
