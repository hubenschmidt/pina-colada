package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

var ErrContactNotFound = errors.New("contact not found")
var ErrContactNameRequired = errors.New("first_name and last_name are required")

// ContactService handles contact business logic
type ContactService struct {
	contactRepo *repositories.ContactRepository
}

// NewContactService creates a new contact service
func NewContactService(contactRepo *repositories.ContactRepository) *ContactService {
	return &ContactService{contactRepo: contactRepo}
}

// GetContacts returns paginated contacts
func (s *ContactService) GetContacts(page, pageSize int, orderBy, order, search string, tenantID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	result, err := s.contactRepo.FindAll(params, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.ContactResponse, len(result.Items))
	for i := range result.Items {
		items[i] = serializers.ContactToResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// GetContact returns a contact by ID
func (s *ContactService) GetContact(id int64) (*serializers.ContactResponse, error) {
	contact, err := s.contactRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if contact == nil {
		return nil, ErrContactNotFound
	}

	resp := serializers.ContactToResponse(contact)
	return &resp, nil
}

// CreateContact creates a new contact
func (s *ContactService) CreateContact(input schemas.ContactCreate, userID int64) (*serializers.ContactResponse, error) {
	if input.FirstName == "" || input.LastName == "" {
		return nil, ErrContactNameRequired
	}

	repoInput := repositories.ContactCreateInput{
		UserID:     userID,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Email:      input.Email,
		Phone:      input.Phone,
		Title:      input.Title,
		Department: input.Department,
		Role:       input.Role,
		Notes:      input.Notes,
		IsPrimary:  input.IsPrimary,
		AccountIDs: input.AccountIDs,
	}

	contactID, err := s.contactRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}

	return s.GetContact(contactID)
}

// UpdateContact updates an existing contact
func (s *ContactService) UpdateContact(id int64, input schemas.ContactUpdate, userID int64) (*serializers.ContactResponse, error) {
	contact, err := s.contactRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if contact == nil {
		return nil, ErrContactNotFound
	}

	if err := s.applyContactUpdates(id, input, userID); err != nil {
		return nil, err
	}

	if err := s.updateAccountLinks(id, input, contact.IsPrimary); err != nil {
		return nil, err
	}

	return s.GetContact(id)
}

func (s *ContactService) applyContactUpdates(id int64, input schemas.ContactUpdate, userID int64) error {
	updates := buildContactUpdates(input, userID)
	if len(updates) <= 1 {
		return nil
	}
	return s.contactRepo.Update(id, updates)
}

func (s *ContactService) updateAccountLinks(id int64, input schemas.ContactUpdate, currentIsPrimary bool) error {
	if input.AccountIDs == nil {
		return nil
	}
	isPrimary := resolvePrimaryFlag(input.IsPrimary, currentIsPrimary)
	return s.contactRepo.LinkToAccounts(id, input.AccountIDs, isPrimary)
}

func resolvePrimaryFlag(inputFlag *bool, current bool) bool {
	if inputFlag != nil {
		return *inputFlag
	}
	return current
}

// DeleteContact deletes a contact
func (s *ContactService) DeleteContact(id int64) error {
	contact, err := s.contactRepo.FindByID(id)
	if err != nil {
		return err
	}
	if contact == nil {
		return ErrContactNotFound
	}

	return s.contactRepo.Delete(id)
}

// SearchContacts searches contacts by name or email
func (s *ContactService) SearchContacts(query string, limit int) ([]serializers.ContactResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	contacts, err := s.contactRepo.Search(query, limit)
	if err != nil {
		return nil, err
	}

	results := make([]serializers.ContactResponse, len(contacts))
	for i := range contacts {
		results[i] = serializers.ContactToResponse(&contacts[i])
	}

	return results, nil
}

func buildContactUpdates(input schemas.ContactUpdate, userID int64) map[string]interface{} {
	updates := map[string]interface{}{
		"updated_by": userID,
	}

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
	if input.Department != nil {
		updates["department"] = *input.Department
	}
	if input.Role != nil {
		updates["role"] = *input.Role
	}
	if input.Notes != nil {
		updates["notes"] = *input.Notes
	}
	if input.IsPrimary != nil {
		updates["is_primary"] = *input.IsPrimary
	}

	return updates
}
