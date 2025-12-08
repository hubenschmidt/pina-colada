package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

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

	return &PaginatedResult[models.Contact]{
		Items:      contacts,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindByID returns a contact by ID
func (r *ContactRepository) FindByID(id int64) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.First(&contact, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &contact, nil
}

// Create creates a new contact
func (r *ContactRepository) Create(contact *models.Contact) error {
	return r.db.Create(contact).Error
}

// Update updates a contact
func (r *ContactRepository) Update(contact *models.Contact, updates map[string]interface{}) error {
	return r.db.Model(contact).Updates(updates).Error
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
