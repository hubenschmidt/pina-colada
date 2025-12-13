package services

import (
	"time"

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

func formatLeadDisplayDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("01/02/2006")
}

func formatLeadDisplayDatePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("01/02/2006")
}

func opportunityToResponse(opp *repositories.OpportunityWithLead) serializers.OpportunityResponse {
	resp := serializers.OpportunityResponse{
		ID:                 opp.ID,
		Account:            opp.AccountName,
		OpportunityName:    opp.OpportunityName,
		EstimatedValue:     opp.EstimatedValue,
		Probability:        opp.Probability,
		Status:             opp.StatusName,
		Description:        opp.Description,
		CreatedAt:          opp.CreatedAt,
		UpdatedAt:          opp.UpdatedAt,
		FormattedUpdatedAt: formatLeadDisplayDate(opp.UpdatedAt),
	}

	if opp.ExpectedCloseDate != nil {
		s := opp.ExpectedCloseDate.Format("2006-01-02")
		resp.ExpectedCloseDate = &s
		resp.FormattedExpectedCloseDate = formatLeadDisplayDatePtr(opp.ExpectedCloseDate)
	}

	return resp
}

func partnershipToResponse(p *repositories.PartnershipWithLead) serializers.PartnershipResponse {
	resp := serializers.PartnershipResponse{
		ID:                 p.ID,
		Account:            p.AccountName,
		PartnershipName:    p.PartnershipName,
		PartnershipType:    p.PartnershipType,
		Status:             p.StatusName,
		Description:        p.Description,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
		FormattedUpdatedAt: formatLeadDisplayDate(p.UpdatedAt),
	}

	if p.StartDate != nil {
		s := p.StartDate.Format("2006-01-02")
		resp.StartDate = &s
		resp.FormattedStartDate = formatLeadDisplayDatePtr(p.StartDate)
	}

	if p.EndDate != nil {
		s := p.EndDate.Format("2006-01-02")
		resp.EndDate = &s
		resp.FormattedEndDate = formatLeadDisplayDatePtr(p.EndDate)
	}

	return resp
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
		items[i] = opportunityToResponse(&result.Items[i])
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

	resp := opportunityToResponse(opp)
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
		items[i] = partnershipToResponse(&result.Items[i])
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

	resp := partnershipToResponse(partnership)
	return &resp, nil
}
