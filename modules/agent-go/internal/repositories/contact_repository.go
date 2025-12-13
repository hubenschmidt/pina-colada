package repositories

import (
	"errors"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// ContactCreateInput contains data needed to create a contact
type ContactCreateInput struct {
	UserID     int64
	FirstName  string
	LastName   string
	Email      *string
	Phone      *string
	Title      *string
	Department *string
	Role       *string
	Notes      *string
	IsPrimary  bool
	AccountIDs []int64
}

// ContactRepository handles contact data access
type ContactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// FindAll returns paginated contacts
func (r *ContactRepository) FindAll(params PaginationParams, tenantID *int64) (*PaginatedResult[models.Contact], error) {
	var contacts []models.Contact
	var totalCount int64

	query := r.db.Model(&models.Contact{})

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(
			`LOWER("Contact".first_name) LIKE LOWER(?) OR LOWER("Contact".last_name) LIKE LOWER(?) OR LOWER("Contact".email) LIKE LOWER(?)`,
			searchTerm, searchTerm, searchTerm,
		)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}

	r.loadContactAccounts(contacts)

	return &PaginatedResult[models.Contact]{
		Items:      contacts,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindByID returns a contact by ID with linked accounts
func (r *ContactRepository) FindByID(id int64) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.First(&contact, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Manually load accounts via Contact_Account join table
	var accountIDs []int64
	r.db.Table(`"Contact_Account"`).
		Select("account_id").
		Where("contact_id = ?", id).
		Pluck("account_id", &accountIDs)

	if len(accountIDs) > 0 {
		var accounts []models.Account
		r.db.Preload("Organizations").
			Preload("Individuals").
			Where("id IN ?", accountIDs).
			Find(&accounts)
		contact.Accounts = accounts
	}

	return &contact, nil
}

// Create creates a new contact and links to accounts
func (r *ContactRepository) Create(input ContactCreateInput) (int64, error) {
	var contactID int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		contact := &models.Contact{
			FirstName:  input.FirstName,
			LastName:   input.LastName,
			Email:      input.Email,
			Phone:      input.Phone,
			Title:      input.Title,
			Department: input.Department,
			Role:       input.Role,
			Notes:      input.Notes,
			IsPrimary:  input.IsPrimary,
			CreatedBy:  input.UserID,
			UpdatedBy:  input.UserID,
		}
		if err := tx.Create(contact).Error; err != nil {
			return err
		}
		contactID = contact.ID

		// Link to accounts if provided
		for _, accountID := range input.AccountIDs {
			ca := models.ContactAccount{
				ContactID: contact.ID,
				AccountID: accountID,
				IsPrimary: input.IsPrimary,
			}
			if err := tx.Create(&ca).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return contactID, err
}

// Update updates a contact by ID
func (r *ContactRepository) Update(id int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Contact{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a contact by ID
func (r *ContactRepository) Delete(id int64) error {
	return r.db.Delete(&models.Contact{}, id).Error
}

// Search searches contacts by name or email
func (r *ContactRepository) Search(query string, limit int) ([]models.Contact, error) {
	var contacts []models.Contact
	searchTerm := "%" + query + "%"

	err := r.db.Model(&models.Contact{}).
		Where(
			`LOWER(first_name) LIKE LOWER(?) OR LOWER(last_name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)`,
			searchTerm, searchTerm, searchTerm,
		).
		Limit(limit).
		Find(&contacts).Error

	return contacts, err
}

// LinkToAccounts links a contact to accounts via Contact_Account
func (r *ContactRepository) LinkToAccounts(contactID int64, accountIDs []int64, isPrimary bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove existing links
		if err := tx.Where("contact_id = ?", contactID).Delete(&models.ContactAccount{}).Error; err != nil {
			return err
		}

		// Add new links
		for _, accountID := range accountIDs {
			ca := models.ContactAccount{
				ContactID: contactID,
				AccountID: accountID,
				IsPrimary: isPrimary,
			}
			if err := tx.Create(&ca).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ContactRepository) loadContactAccounts(contacts []models.Contact) {
	if len(contacts) == 0 {
		return
	}

	contactIDs := make([]int64, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	var links []models.ContactAccount
	r.db.Where("contact_id IN ?", contactIDs).Find(&links)

	accountIDs := extractUniqueAccountIDs(links)
	accountMap := r.loadAccountMap(accountIDs)
	contactAccounts := buildContactAccountsMap(links, accountMap)

	for i := range contacts {
		contacts[i].Accounts = contactAccounts[contacts[i].ID]
	}
}

func extractUniqueAccountIDs(links []models.ContactAccount) []int64 {
	accountIDSet := make(map[int64]bool)
	for _, link := range links {
		accountIDSet[link.AccountID] = true
	}
	accountIDs := make([]int64, 0, len(accountIDSet))
	for id := range accountIDSet {
		accountIDs = append(accountIDs, id)
	}
	return accountIDs
}

func (r *ContactRepository) loadAccountMap(accountIDs []int64) map[int64]models.Account {
	accountMap := make(map[int64]models.Account)
	if len(accountIDs) == 0 {
		return accountMap
	}
	var accounts []models.Account
	r.db.Where("id IN ?", accountIDs).Find(&accounts)
	for _, acc := range accounts {
		accountMap[acc.ID] = acc
	}
	return accountMap
}

func buildContactAccountsMap(links []models.ContactAccount, accountMap map[int64]models.Account) map[int64][]models.Account {
	contactAccounts := make(map[int64][]models.Account)
	for _, link := range links {
		acc, ok := accountMap[link.AccountID]
		if ok {
			contactAccounts[link.ContactID] = append(contactAccounts[link.ContactID], acc)
		}
	}
	return contactAccounts
}
