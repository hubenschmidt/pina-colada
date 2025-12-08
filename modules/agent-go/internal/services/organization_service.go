package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
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

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(id int64, input schemas.OrganizationUpdate, userID int64) (*serializers.OrganizationDetailResponse, error) {
	org, err := s.orgRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("organization not found")
	}

	updates := buildOrgUpdates(input)
	if len(updates) > 0 {
		if err := s.orgRepo.Update(org, updates); err != nil {
			return nil, err
		}
	}

	// Handle industry updates if provided
	if input.IndustryIDs != nil && org.AccountID != nil {
		if err := s.orgRepo.UpdateAccountIndustries(*org.AccountID, input.IndustryIDs); err != nil {
			return nil, err
		}
	}

	return s.GetOrganization(id)
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
		return nil, errors.New("organization not found")
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
