package repositories

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// IndividualRepository handles individual data access
type IndividualRepository struct {
	db *gorm.DB
}

// NewIndividualRepository creates a new individual repository
func NewIndividualRepository(db *gorm.DB) *IndividualRepository {
	return &IndividualRepository{db: db}
}

// FindAll returns paginated individuals
func (r *IndividualRepository) FindAll(params PaginationParams, tenantID *int64) (*PaginatedResult[models.Individual], error) {
	var individuals []models.Individual
	var totalCount int64

	query := r.db.Model(&models.Individual{}).
		Preload("Account").
		Preload("Account.Industries")

	if tenantID != nil {
		query = query.Joins(`INNER JOIN "Account" ON "Account".id = "Individual".account_id`).
			Where(`"Account".tenant_id = ?`, *tenantID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(
			`LOWER("Individual".first_name) LIKE LOWER(?) OR LOWER("Individual".last_name) LIKE LOWER(?) OR LOWER("Individual".email) LIKE LOWER(?)`,
			searchTerm, searchTerm, searchTerm,
		)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&individuals).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[models.Individual]{
		Items:      individuals,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindByID returns an individual by ID
func (r *IndividualRepository) FindByID(id int64) (*models.Individual, error) {
	var ind models.Individual
	err := r.db.
		Preload("Account").
		Preload("Account.Industries").
		Preload("Account.Contacts").
		Preload("Account.OutgoingRelationships.ToAccount.Organizations").
		Preload("Account.OutgoingRelationships.ToAccount.Individuals").
		Preload("Account.IncomingRelationships.FromAccount.Organizations").
		Preload("Account.IncomingRelationships.FromAccount.Individuals").
		First(&ind, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ind, nil
}

// Create creates a new individual
func (r *IndividualRepository) Create(ind *models.Individual) error {
	return r.db.Create(ind).Error
}

// Update updates an individual
func (r *IndividualRepository) Update(ind *models.Individual, updates map[string]interface{}) error {
	return r.db.Model(&models.Individual{ID: ind.ID}).Updates(updates).Error
}

// Delete deletes an individual by ID
func (r *IndividualRepository) Delete(id int64) error {
	return r.db.Delete(&models.Individual{}, id).Error
}

// IndContactInput holds data for creating a contact for an individual
type IndContactInput struct {
	FirstName  *string
	LastName   *string
	Email      *string
	Phone      *string
	Title      *string
	Department *string
	Role       *string
	IsPrimary  bool
	Notes      *string
}

// CreateContactForAccount creates a contact and links it to an account
func (r *IndividualRepository) CreateContactForAccount(accountID int64, input IndContactInput, userID int64) (*models.Contact, error) {
	var firstName, lastName string
	if input.FirstName != nil {
		firstName = *input.FirstName
	}
	if input.LastName != nil {
		lastName = *input.LastName
	}

	contact := &models.Contact{
		FirstName: firstName,
		LastName:  lastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Title:     input.Title,
		CreatedBy: userID,
		UpdatedBy: userID,
	}

	if err := r.db.Create(contact).Error; err != nil {
		return nil, err
	}

	// Link to account
	link := &models.ContactAccount{
		ContactID: contact.ID,
		AccountID: accountID,
		IsPrimary: input.IsPrimary,
	}
	if err := r.db.Create(link).Error; err != nil {
		return nil, err
	}

	return contact, nil
}

// Search searches individuals by name or email
func (r *IndividualRepository) Search(query string, tenantID *int64, limit int) ([]models.Individual, error) {
	var individuals []models.Individual
	searchTerm := "%" + query + "%"

	// Also search by concatenated first+last name for queries like "William Hubenschmidt"
	q := r.db.Model(&models.Individual{}).
		Preload("Account").
		Where(
			`LOWER("Individual".first_name) LIKE LOWER(?) OR LOWER("Individual".last_name) LIKE LOWER(?) OR LOWER("Individual".email) LIKE LOWER(?) OR LOWER(CONCAT("Individual".first_name, ' ', "Individual".last_name)) LIKE LOWER(?)`,
			searchTerm, searchTerm, searchTerm, searchTerm,
		)

	if tenantID != nil {
		q = q.Joins(`INNER JOIN "Account" ON "Account".id = "Individual".account_id`).
			Where(`"Account".tenant_id = ?`, *tenantID)
	}

	err := q.Limit(limit).Find(&individuals).Error
	return individuals, err
}

// FindOrCreateByName finds an individual by first+last name or creates it with an Account
func (r *IndividualRepository) FindOrCreateByName(firstName, lastName string, tenantID *int64, userID int64) (*models.Individual, error) {
	var ind models.Individual
	err := r.db.Preload("Account").
		Where(`LOWER("Individual".first_name) = LOWER(?) AND LOWER("Individual".last_name) = LOWER(?)`, firstName, lastName).
		First(&ind).Error

	if err == nil {
		return &ind, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create Account first
	fullName := firstName + " " + lastName
	account := &models.Account{
		TenantID:  tenantID,
		Name:      fullName,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := r.db.Create(account).Error; err != nil {
		return nil, err
	}

	// Create Individual linked to Account
	newInd := &models.Individual{
		AccountID: &account.ID,
		FirstName: firstName,
		LastName:  lastName,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := r.db.Create(newInd).Error; err != nil {
		return nil, err
	}

	newInd.Account = account
	return newInd, nil
}

// IndividualCreateInput holds data for creating an individual
type IndividualCreateInput struct {
	FirstName       string
	LastName        string
	Email           *string
	Phone           *string
	LinkedInURL     *string
	Title           *string
	Description     *string
	AccountID       *int64
	TwitterURL      *string
	GithubURL       *string
	Bio             *string
	SeniorityLevel  *string
	Department      *string
	IsDecisionMaker *bool
	ReportsToID     *int64
	TenantID        *int64
	UserID          int64
	IndustryIDs     []int64
	ProjectIDs      []int64
}

// CreateIndividual creates a new individual with account
func (r *IndividualRepository) CreateIndividual(input IndividualCreateInput) (*models.Individual, error) {
	var accountID *int64

	if input.AccountID != nil {
		accountID = input.AccountID
	} else {
		// Create a new account for this individual
		fullName := input.FirstName + " " + input.LastName
		account := &models.Account{
			TenantID:  input.TenantID,
			Name:      fullName,
			CreatedBy: input.UserID,
			UpdatedBy: input.UserID,
		}
		if err := r.db.Create(account).Error; err != nil {
			return nil, err
		}
		accountID = &account.ID

		// Add industries to account if provided
		if len(input.IndustryIDs) > 0 {
			for _, industryID := range input.IndustryIDs {
				link := &models.AccountIndustry{
					AccountID:  account.ID,
					IndustryID: industryID,
				}
				r.db.Create(link)
			}
		}
	}

	ind := &models.Individual{
		AccountID:       accountID,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Email:           input.Email,
		Phone:           input.Phone,
		LinkedInURL:     input.LinkedInURL,
		Title:           input.Title,
		Description:     input.Description,
		TwitterURL:      input.TwitterURL,
		GithubURL:       input.GithubURL,
		Bio:             input.Bio,
		SeniorityLevel:  input.SeniorityLevel,
		Department:      input.Department,
		IsDecisionMaker: input.IsDecisionMaker,
		ReportsToID:     input.ReportsToID,
		CreatedBy:       input.UserID,
		UpdatedBy:       input.UserID,
	}

	if err := r.db.Create(ind).Error; err != nil {
		return nil, err
	}

	// Link to projects via account if provided
	if len(input.ProjectIDs) > 0 && accountID != nil {
		for _, projectID := range input.ProjectIDs {
			link := &models.AccountProject{
				ProjectID: projectID,
				AccountID: *accountID,
			}
			r.db.Create(link)
		}
	}

	// Reload with associations
	return r.FindByID(ind.ID)
}

// GetContactsForIndividual returns contacts for an individual via their account
func (r *IndividualRepository) GetContactsForIndividual(individualID int64) ([]models.Contact, error) {
	var ind models.Individual
	err := r.db.First(&ind, individualID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if ind.AccountID == nil {
		return []models.Contact{}, nil
	}

	var contacts []models.Contact
	err = r.db.
		Joins(`INNER JOIN "ContactAccount" ON "ContactAccount".contact_id = "Contact".id`).
		Where(`"ContactAccount".account_id = ?`, *ind.AccountID).
		Find(&contacts).Error

	return contacts, err
}

// UpdateContact updates a contact
func (r *IndividualRepository) UpdateContact(contactID int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Contact{ID: contactID}).Updates(updates).Error
}

// DeleteContactFromAccount deletes a contact-account link and contact if orphaned
func (r *IndividualRepository) DeleteContactFromAccount(individualID, contactID int64) error {
	var ind models.Individual
	if err := r.db.First(&ind, individualID).Error; err != nil {
		return err
	}
	if ind.AccountID == nil {
		return gorm.ErrRecordNotFound
	}

	// Delete the link
	if err := r.db.Where("contact_id = ? AND account_id = ?", contactID, *ind.AccountID).
		Delete(&models.ContactAccount{}).Error; err != nil {
		return err
	}

	// Check if contact is orphaned
	var count int64
	r.db.Model(&models.ContactAccount{}).Where("contact_id = ?", contactID).Count(&count)
	if count == 0 {
		return r.db.Delete(&models.Contact{}, contactID).Error
	}

	return nil
}

// SignalCreateInput holds data for creating a signal
type SignalCreateInput struct {
	SignalType     string
	Headline       string
	Description    *string
	SignalDate     *time.Time
	Source         *string
	SourceURL      *string
	Sentiment      *string
	RelevanceScore *decimal.Decimal
}

// GetSignalsForIndividual returns signals for an individual via their account
func (r *IndividualRepository) GetSignalsForIndividual(individualID int64, signalType *string, limit int) ([]models.Signal, error) {
	var ind models.Individual
	err := r.db.First(&ind, individualID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if ind.AccountID == nil {
		return []models.Signal{}, nil
	}

	query := r.db.Where("account_id = ?", *ind.AccountID)
	if signalType != nil && *signalType != "" {
		query = query.Where("signal_type = ?", *signalType)
	}

	var signals []models.Signal
	err = query.Order("created_at DESC").Limit(limit).Find(&signals).Error
	return signals, err
}

// CreateSignalForIndividual creates a signal for an individual's account
func (r *IndividualRepository) CreateSignalForIndividual(individualID int64, input SignalCreateInput) (*models.Signal, error) {
	var ind models.Individual
	if err := r.db.First(&ind, individualID).Error; err != nil {
		return nil, err
	}

	if ind.AccountID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	signal := &models.Signal{
		AccountID:      *ind.AccountID,
		SignalType:     input.SignalType,
		Headline:       input.Headline,
		Description:    input.Description,
		SignalDate:     input.SignalDate,
		Source:         input.Source,
		SourceURL:      input.SourceURL,
		Sentiment:      input.Sentiment,
		RelevanceScore: input.RelevanceScore,
	}

	if err := r.db.Create(signal).Error; err != nil {
		return nil, err
	}

	return signal, nil
}

// DeleteSignal deletes a signal
func (r *IndividualRepository) DeleteSignal(individualID, signalID int64) error {
	var ind models.Individual
	if err := r.db.First(&ind, individualID).Error; err != nil {
		return err
	}

	if ind.AccountID == nil {
		return gorm.ErrRecordNotFound
	}

	return r.db.Where("id = ? AND account_id = ?", signalID, *ind.AccountID).
		Delete(&models.Signal{}).Error
}

// GetContactByID returns a contact by ID
func (r *IndividualRepository) GetContactByID(contactID int64) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.First(&contact, contactID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &contact, nil
}
