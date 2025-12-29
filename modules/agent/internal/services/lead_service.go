package services

import (
	"errors"
	"time"

	"agent/internal/repositories"
	"agent/internal/schemas"
	"agent/internal/serializers"
)

var (
	ErrOpportunityNotFound = errors.New("opportunity not found")
	ErrPartnershipNotFound = errors.New("partnership not found")
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

// CreateOpportunity creates a new opportunity
func (s *LeadService) CreateOpportunity(input schemas.OpportunityCreate, tenantID *int64, userID int64) (*serializers.OpportunityResponse, error) {
	repoInput := repositories.OpportunityCreateInput{
		TenantID:          tenantID,
		UserID:            userID,
		Source:            input.Source,
		OpportunityName:   input.OpportunityName,
		EstimatedValue:    input.EstimatedValue,
		Probability:       input.Probability,
		ExpectedCloseDate: input.ExpectedCloseDate,
		Description:       input.Description,
		ProjectIDs:        input.ProjectIDs,
	}
	id, err := s.leadRepo.CreateOpportunity(repoInput)
	if err != nil {
		return nil, err
	}
	return s.GetOpportunity(id)
}

// UpdateOpportunity updates an existing opportunity
func (s *LeadService) UpdateOpportunity(id int64, input schemas.OpportunityUpdate, userID int64) (*serializers.OpportunityResponse, error) {
	repoInput := repositories.OpportunityUpdateInput{
		UserID:            userID,
		OpportunityName:   input.OpportunityName,
		EstimatedValue:    input.EstimatedValue,
		Probability:       input.Probability,
		ExpectedCloseDate: input.ExpectedCloseDate,
		Description:       input.Description,
	}
	err := s.leadRepo.UpdateOpportunity(id, repoInput)
	if err != nil {
		return nil, err
	}
	return s.GetOpportunity(id)
}

// DeleteOpportunity deletes an opportunity
func (s *LeadService) DeleteOpportunity(id int64) error {
	return s.leadRepo.DeleteOpportunity(id)
}

// CreatePartnership creates a new partnership
func (s *LeadService) CreatePartnership(input schemas.PartnershipCreate, tenantID *int64, userID int64) (*serializers.PartnershipResponse, error) {
	repoInput := repositories.PartnershipCreateInput{
		TenantID:        tenantID,
		UserID:          userID,
		Source:          input.Source,
		PartnershipName: input.PartnershipName,
		PartnershipType: input.PartnershipType,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		Description:     input.Description,
		ProjectIDs:      input.ProjectIDs,
	}
	id, err := s.leadRepo.CreatePartnership(repoInput)
	if err != nil {
		return nil, err
	}
	return s.GetPartnership(id)
}

// UpdatePartnership updates an existing partnership
func (s *LeadService) UpdatePartnership(id int64, input schemas.PartnershipUpdate, userID int64) (*serializers.PartnershipResponse, error) {
	repoInput := repositories.PartnershipUpdateInput{
		UserID:          userID,
		PartnershipName: input.PartnershipName,
		PartnershipType: input.PartnershipType,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		Description:     input.Description,
	}
	err := s.leadRepo.UpdatePartnership(id, repoInput)
	if err != nil {
		return nil, err
	}
	return s.GetPartnership(id)
}

// DeletePartnership deletes a partnership
func (s *LeadService) DeletePartnership(id int64) error {
	return s.leadRepo.DeletePartnership(id)
}
