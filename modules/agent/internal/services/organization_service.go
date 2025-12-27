package services

import (
	"errors"

	"github.com/shopspring/decimal"

	"github.com/pina-colada-co/agent-go/internal/lib"
	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

// ErrOrganizationNotFound is returned when an organization is not found
var ErrOrganizationNotFound = errors.New("organization not found")

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
		return nil, ErrOrganizationNotFound
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
		return ErrOrganizationNotFound
	}

	return s.orgRepo.Delete(id)
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(id int64, input schemas.OrganizationUpdate, userID int64) (*serializers.OrganizationDetailResponse, error) {
	org, err := s.orgRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	if err := s.applyOrgUpdates(org, input); err != nil {
		return nil, err
	}

	if err := s.updateOrgIndustries(org.AccountID, input.IndustryIDs); err != nil {
		return nil, err
	}

	return s.GetOrganization(id)
}

func (s *OrganizationService) applyOrgUpdates(org *models.Organization, input schemas.OrganizationUpdate) error {
	updates := buildOrgUpdates(input)
	if len(updates) == 0 {
		return nil
	}
	return s.orgRepo.Update(org, updates)
}

func (s *OrganizationService) updateOrgIndustries(accountID *int64, industryIDs []int64) error {
	if industryIDs == nil || accountID == nil {
		return nil
	}
	return s.orgRepo.UpdateAccountIndustries(*accountID, industryIDs)
}

func buildOrgUpdates(input schemas.OrganizationUpdate) map[string]interface{} {
	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Website != nil {
		updates["website"] = *input.Website
	}
	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.EmployeeCount != nil {
		updates["employee_count"] = *input.EmployeeCount
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.LinkedInURL != nil {
		updates["linkedin_url"] = *input.LinkedInURL
	}
	if input.FoundingYear != nil {
		updates["founded"] = *input.FoundingYear
	}
	if input.HeadquartersCity != nil {
		updates["headquarters"] = *input.HeadquartersCity
	}
	return updates
}

// AddContactToOrganization adds a contact to an organization
func (s *OrganizationService) AddContactToOrganization(orgID int64, input schemas.OrgContactCreate, userID int64, tenantID int64) (*serializers.ContactResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil || org.AccountID == nil {
		return nil, ErrOrganizationNotFound
	}

	repoInput := repositories.OrgContactInput{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Title:     input.Title,
		IsPrimary: input.IsPrimary,
	}

	contact, err := s.orgRepo.CreateContactForAccount(*org.AccountID, repoInput, userID)
	if err != nil {
		return nil, err
	}

	return &serializers.ContactResponse{
		ID:        contact.ID,
		FirstName: contact.FirstName,
		LastName:  contact.LastName,
		Email:     contact.Email,
		Phone:     contact.Phone,
		Title:     contact.Title,
	}, nil
}

// GetOrganizationContacts returns contacts for an organization
func (s *OrganizationService) GetOrganizationContacts(orgID int64) ([]serializers.ContactResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil || org.AccountID == nil {
		return nil, ErrOrganizationNotFound
	}

	contacts, err := s.orgRepo.GetContactsForAccount(*org.AccountID)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.ContactResponse, len(contacts))
	for i, c := range contacts {
		results[i] = serializers.ContactResponse{
			ID:        c.ID,
			FirstName: c.FirstName,
			LastName:  c.LastName,
			Email:     c.Email,
			Phone:     c.Phone,
			Title:     c.Title,
		}
	}
	return results, nil
}

// UpdateOrganizationContact updates a contact for an organization
func (s *OrganizationService) UpdateOrganizationContact(orgID int64, contactID int64, input schemas.OrgContactUpdate, userID int64) (*serializers.ContactResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil || org.AccountID == nil {
		return nil, ErrOrganizationNotFound
	}

	updates := buildOrgContactUpdates(input, userID)
	var updateErr error
	if len(updates) > 0 {
		updateErr = s.orgRepo.UpdateContact(contactID, updates)
	}
	if updateErr != nil {
		return nil, updateErr
	}

	// For now return minimal response
	return &serializers.ContactResponse{ID: contactID}, nil
}

func buildOrgContactUpdates(input schemas.OrgContactUpdate, userID int64) map[string]interface{} {
	updates := map[string]interface{}{"updated_by": userID}
	if input.FirstName != nil {
		updates["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		updates["last_name"] = *input.LastName
	}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	return updates
}

// DeleteOrganizationContact deletes a contact from an organization
func (s *OrganizationService) DeleteOrganizationContact(orgID int64, contactID int64) error {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return err
	}
	if org == nil || org.AccountID == nil {
		return ErrOrganizationNotFound
	}

	return s.orgRepo.DeleteContactFromAccount(*org.AccountID, contactID)
}

// GetOrganizationTechnologies returns technologies for an organization
func (s *OrganizationService) GetOrganizationTechnologies(orgID int64) ([]serializers.OrgTechnologyResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	techs, err := s.orgRepo.GetTechnologiesForOrg(orgID)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.OrgTechnologyResponse, len(techs))
	for i, t := range techs {
		results[i] = serializers.OrgTechnologyResponse{
			OrganizationID: t.OrganizationID,
			TechnologyID:   t.TechnologyID,
			Source:         t.Source,
			DetectedAt:     t.DetectedAt.Format("2006-01-02T15:04:05Z"),
		}
		if t.Confidence != nil {
			f, _ := t.Confidence.Float64()
			results[i].Confidence = &f
		}
	}
	return results, nil
}

// AddTechnologyToOrganization adds a technology to an organization
func (s *OrganizationService) AddTechnologyToOrganization(orgID int64, input schemas.OrgTechnologyCreate) error {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return err
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	return s.orgRepo.AddTechnologyToOrg(orgID, input.TechnologyID, input.Source, input.Confidence)
}

// RemoveTechnologyFromOrganization removes a technology from an organization
func (s *OrganizationService) RemoveTechnologyFromOrganization(orgID int64, techID int64) error {
	return s.orgRepo.RemoveTechnologyFromOrg(orgID, techID)
}

// GetOrganizationFundingRounds returns funding rounds for an organization
func (s *OrganizationService) GetOrganizationFundingRounds(orgID int64) ([]serializers.FundingRoundResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	rounds, err := s.orgRepo.GetFundingRoundsForOrg(orgID)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.FundingRoundResponse, len(rounds))
	for i, r := range rounds {
		results[i] = serializers.FundingRoundResponse{
			ID:             r.ID,
			OrganizationID: r.OrganizationID,
			RoundType:      r.RoundType,
			Amount:         r.Amount,
			LeadInvestor:   r.LeadInvestor,
			SourceURL:      r.SourceURL,
		}
		if r.AnnouncedDate != nil {
			d := r.AnnouncedDate.Format("2006-01-02")
			results[i].AnnouncedDate = &d
		}
	}
	return results, nil
}

// CreateOrganizationFundingRound creates a funding round for an organization
func (s *OrganizationService) CreateOrganizationFundingRound(orgID int64, input schemas.FundingRoundCreate) (*serializers.FundingRoundResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	round := &repositories.FundingRoundInput{
		OrganizationID: orgID,
		RoundType:      input.RoundType,
		Amount:         input.Amount,
		LeadInvestor:   input.LeadInvestor,
		SourceURL:      input.SourceURL,
		AnnouncedDate:  input.AnnouncedDate,
	}

	created, err := s.orgRepo.CreateFundingRoundFromInput(round)
	if err != nil {
		return nil, err
	}

	resp := &serializers.FundingRoundResponse{
		ID:             created.ID,
		OrganizationID: created.OrganizationID,
		RoundType:      created.RoundType,
		Amount:         created.Amount,
		LeadInvestor:   created.LeadInvestor,
		SourceURL:      created.SourceURL,
	}
	if created.AnnouncedDate != nil {
		d := created.AnnouncedDate.Format("2006-01-02")
		resp.AnnouncedDate = &d
	}
	return resp, nil
}

// DeleteOrganizationFundingRound deletes a funding round
func (s *OrganizationService) DeleteOrganizationFundingRound(orgID int64, roundID int64) error {
	return s.orgRepo.DeleteFundingRound(roundID)
}

// GetOrganizationSignals returns signals for an organization
func (s *OrganizationService) GetOrganizationSignals(orgID int64, signalType *string, limit int) ([]serializers.SignalResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil || org.AccountID == nil {
		return nil, ErrOrganizationNotFound
	}

	if limit <= 0 {
		limit = 20
	}

	signals, err := s.orgRepo.GetSignalsForAccount(*org.AccountID, signalType, limit)
	if err != nil {
		return nil, err
	}

	return serializers.SignalsToResponse(signals), nil
}

// CreateOrganizationSignal creates a signal for an organization
func (s *OrganizationService) CreateOrganizationSignal(orgID int64, input schemas.SignalCreate, userID int64) (*serializers.SignalResponse, error) {
	org, err := s.orgRepo.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	if org == nil || org.AccountID == nil {
		return nil, ErrOrganizationNotFound
	}

	signalInput := repositories.SignalCreateInput{
		SignalType:  input.SignalType,
		Headline:    input.Headline,
		Description: input.Description,
		Source:      input.Source,
		SourceURL:   input.SourceURL,
		Sentiment:   input.Sentiment,
	}
	signalInput.SignalDate = lib.ParseDateString(input.SignalDate)
	if input.RelevanceScore != nil {
		dec := decimal.NewFromFloat(*input.RelevanceScore)
		signalInput.RelevanceScore = &dec
	}

	created, err := s.orgRepo.CreateSignalForAccount(*org.AccountID, signalInput)
	if err != nil {
		return nil, err
	}

	resp := serializers.SignalToResponse(created)
	return &resp, nil
}

// DeleteOrganizationSignal deletes a signal
func (s *OrganizationService) DeleteOrganizationSignal(orgID int64, signalID int64) error {
	return s.orgRepo.DeleteSignal(signalID)
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(input schemas.OrganizationCreate, userID int64, tenantID *int64) (*serializers.OrganizationDetailResponse, error) {
	// Use existing account if provided
	accountID := int64(0)
	if input.AccountID != nil {
		accountID = *input.AccountID
	}

	// Create new account if not provided
	if accountID == 0 {
		account, err := s.orgRepo.CreateAccount(input.Name, tenantID, userID)
		if err != nil {
			return nil, err
		}
		accountID = account.ID
	}

	org, err := s.orgRepo.CreateOrganization(repositories.OrganizationCreateInput{
		AccountID:            accountID,
		Name:                 input.Name,
		Website:              input.Website,
		Phone:                input.Phone,
		EmployeeCount:        input.EmployeeCount,
		EmployeeCountRangeID: input.EmployeeCountRangeID,
		FundingStageID:       input.FundingStageID,
		Description:          input.Description,
		RevenueRangeID:       input.RevenueRangeID,
		FoundingYear:         input.FoundingYear,
		HeadquartersCity:     input.HeadquartersCity,
		HeadquartersState:    input.HeadquartersState,
		HeadquartersCountry:  input.HeadquartersCountry,
		CompanyType:          input.CompanyType,
		LinkedInURL:          input.LinkedInURL,
		CrunchbaseURL:        input.CrunchbaseURL,
		UserID:               userID,
	})
	if err != nil {
		return nil, err
	}

	// Handle industry associations if provided
	if len(input.IndustryIDs) > 0 {
		s.orgRepo.UpdateAccountIndustries(accountID, input.IndustryIDs)
	}

	return s.GetOrganization(org.ID)
}
