package services

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

// ErrIndividualNotFound is returned when an individual is not found
var ErrIndividualNotFound = errors.New("individual not found")

// ErrIndividualNoAccount is returned when an individual has no account
var ErrIndividualNoAccount = errors.New("individual has no account")

// ErrIndividualContactNotFound is returned when a contact is not found
var ErrIndividualContactNotFound = errors.New("contact not found")

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
		return nil, ErrIndividualNotFound
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
		return ErrIndividualNotFound
	}

	return s.indRepo.Delete(id)
}

// IndividualUpdateInput holds fields for updating an individual
type IndividualUpdateInput struct {
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	LinkedInURL    *string `json:"linkedin_url"`
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	TwitterURL     *string `json:"twitter_url"`
	GithubURL      *string `json:"github_url"`
	Bio            *string `json:"bio"`
	SeniorityLevel *string `json:"seniority_level"`
	Department     *string `json:"department"`
	IsDecisionMaker *bool  `json:"is_decision_maker"`
	IndustryIDs    []int64 `json:"industry_ids"`
}

// UpdateIndividual updates an individual
func (s *IndividualService) UpdateIndividual(id int64, input IndividualUpdateInput, userID int64) (*serializers.IndividualDetailResponse, error) {
	ind, err := s.indRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if ind == nil {
		return nil, ErrIndividualNotFound
	}

	updates := buildIndividualUpdates(input, userID)
	if len(updates) > 0 {
		err = s.indRepo.Update(ind, updates)
		if err != nil {
			return nil, err
		}
	}

	// Refetch to get updated data
	ind, err = s.indRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	resp := serializers.IndividualToDetailResponse(ind)
	return &resp, nil
}

// ContactCreateInput holds data for creating a contact for an individual
type ContactCreateInput struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	Title      *string `json:"title"`
	Department *string `json:"department"`
	Role       *string `json:"role"`
	IsPrimary  bool    `json:"is_primary"`
	Notes      *string `json:"notes"`
}

// AddContactToIndividual creates a contact and links it to an individual's account
func (s *IndividualService) AddContactToIndividual(individualID int64, input ContactCreateInput, userID int64) (*serializers.ContactBrief, error) {
	ind, err := s.indRepo.FindByID(individualID)
	if err != nil {
		return nil, err
	}
	if ind == nil {
		return nil, ErrIndividualNotFound
	}
	if ind.AccountID == nil {
		return nil, ErrIndividualNoAccount
	}

	repoInput := repositories.IndContactInput{
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Email:      input.Email,
		Phone:      input.Phone,
		Title:      input.Title,
		Department: input.Department,
		Role:       input.Role,
		IsPrimary:  input.IsPrimary,
		Notes:      input.Notes,
	}
	contact, err := s.indRepo.CreateContactForAccount(*ind.AccountID, repoInput, userID)
	if err != nil {
		return nil, err
	}

	return &serializers.ContactBrief{
		ID:        contact.ID,
		FirstName: contact.FirstName,
		LastName:  contact.LastName,
		Email:     contact.Email,
		Phone:     contact.Phone,
		Title:     contact.Title,
		IsPrimary: input.IsPrimary,
	}, nil
}

func buildIndividualUpdates(input IndividualUpdateInput, userID int64) map[string]interface{} {
	updates := make(map[string]interface{})
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
	if input.LinkedInURL != nil {
		updates["linkedin_url"] = *input.LinkedInURL
	}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.TwitterURL != nil {
		updates["twitter_url"] = *input.TwitterURL
	}
	if input.GithubURL != nil {
		updates["github_url"] = *input.GithubURL
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}
	if input.SeniorityLevel != nil {
		updates["seniority_level"] = *input.SeniorityLevel
	}
	if input.Department != nil {
		updates["department"] = *input.Department
	}
	if input.IsDecisionMaker != nil {
		updates["is_decision_maker"] = *input.IsDecisionMaker
	}
	updates["updated_by"] = userID
	return updates
}

// CreateIndividual creates a new individual
func (s *IndividualService) CreateIndividual(input schemas.IndividualCreate, tenantID *int64, userID int64) (*serializers.IndividualDetailResponse, error) {
	repoInput := repositories.IndividualCreateInput{
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Email:           input.Email,
		Phone:           input.Phone,
		LinkedInURL:     input.LinkedInURL,
		Title:           input.Title,
		Description:     input.Description,
		AccountID:       input.AccountID,
		TwitterURL:      input.TwitterURL,
		GithubURL:       input.GithubURL,
		Bio:             input.Bio,
		SeniorityLevel:  input.SeniorityLevel,
		Department:      input.Department,
		IsDecisionMaker: input.IsDecisionMaker,
		ReportsToID:     input.ReportsToID,
		TenantID:        tenantID,
		UserID:          userID,
		IndustryIDs:     input.IndustryIDs,
		ProjectIDs:      input.ProjectIDs,
	}

	ind, err := s.indRepo.CreateIndividual(repoInput)
	if err != nil {
		return nil, err
	}

	resp := serializers.IndividualToDetailResponse(ind)
	return &resp, nil
}

// GetContacts returns contacts for an individual
func (s *IndividualService) GetContacts(individualID int64) ([]serializers.ContactBrief, error) {
	contacts, err := s.indRepo.GetContactsForIndividual(individualID)
	if err != nil {
		return nil, err
	}
	if contacts == nil {
		return nil, ErrIndividualNotFound
	}

	result := make([]serializers.ContactBrief, len(contacts))
	for i, c := range contacts {
		result[i] = serializers.ContactBrief{
			ID:        c.ID,
			FirstName: c.FirstName,
			LastName:  c.LastName,
			Email:     c.Email,
			Phone:     c.Phone,
			Title:     c.Title,
		}
	}
	return result, nil
}

// UpdateContact updates a contact for an individual
func (s *IndividualService) UpdateContact(individualID, contactID int64, input schemas.IndContactUpdate, userID int64) (*serializers.ContactBrief, error) {
	ind, err := s.indRepo.FindByID(individualID)
	if err != nil {
		return nil, err
	}
	if ind == nil {
		return nil, ErrIndividualNotFound
	}

	updates := buildIndContactUpdates(input, userID)
	if len(updates) > 0 {
		err := s.indRepo.UpdateContact(contactID, updates)
		if err != nil {
			return nil, err
		}
	}

	contact, err := s.indRepo.GetContactByID(contactID)
	if err != nil {
		return nil, err
	}
	if contact == nil {
		return nil, ErrIndividualContactNotFound
	}

	return &serializers.ContactBrief{
		ID:        contact.ID,
		FirstName: contact.FirstName,
		LastName:  contact.LastName,
		Email:     contact.Email,
		Phone:     contact.Phone,
		Title:     contact.Title,
	}, nil
}

func buildIndContactUpdates(input schemas.IndContactUpdate, userID int64) map[string]interface{} {
	updates := make(map[string]interface{})
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
	updates["updated_by"] = userID
	return updates
}

// DeleteContact deletes a contact from an individual
func (s *IndividualService) DeleteContact(individualID, contactID int64) error {
	ind, err := s.indRepo.FindByID(individualID)
	if err != nil {
		return err
	}
	if ind == nil {
		return ErrIndividualNotFound
	}

	return s.indRepo.DeleteContactFromAccount(individualID, contactID)
}

// GetSignals returns signals for an individual
func (s *IndividualService) GetSignals(individualID int64, signalType *string, limit int) ([]serializers.SignalResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	signals, err := s.indRepo.GetSignalsForIndividual(individualID, signalType, limit)
	if err != nil {
		return nil, err
	}
	if signals == nil {
		return nil, ErrIndividualNotFound
	}

	return serializers.SignalsToResponse(signals), nil
}

// CreateSignal creates a signal for an individual
func (s *IndividualService) CreateSignal(individualID int64, input schemas.SignalCreate) (*serializers.SignalResponse, error) {
	repoInput := repositories.SignalCreateInput{
		SignalType:  input.SignalType,
		Headline:    input.Headline,
		Description: input.Description,
		Source:      input.Source,
		SourceURL:   input.SourceURL,
		Sentiment:   input.Sentiment,
	}

	if input.SignalDate != nil {
		t, err := time.Parse("2006-01-02", *input.SignalDate)
		if err == nil {
			repoInput.SignalDate = &t
		}
	}

	if input.RelevanceScore != nil {
		d := decimal.NewFromFloat(*input.RelevanceScore)
		repoInput.RelevanceScore = &d
	}

	signal, err := s.indRepo.CreateSignalForIndividual(individualID, repoInput)
	if err != nil {
		return nil, err
	}

	resp := serializers.SignalToResponse(signal)
	return &resp, nil
}

// DeleteSignal deletes a signal from an individual
func (s *IndividualService) DeleteSignal(individualID, signalID int64) error {
	return s.indRepo.DeleteSignal(individualID, signalID)
}
