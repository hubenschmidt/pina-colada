package repositories

import (
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

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
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

	q := r.db.Model(&models.Individual{}).
		Preload("Account").
		Where(
			`LOWER("Individual".first_name) LIKE LOWER(?) OR LOWER("Individual".last_name) LIKE LOWER(?) OR LOWER("Individual".email) LIKE LOWER(?)`,
			searchTerm, searchTerm, searchTerm,
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
