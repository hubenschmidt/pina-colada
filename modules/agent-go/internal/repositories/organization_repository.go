package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// OrganizationRepository handles organization data access
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// FindAll returns paginated organizations
func (r *OrganizationRepository) FindAll(params PaginationParams, tenantID *int64) (*PaginatedResult[models.Organization], error) {
	var orgs []models.Organization
	var totalCount int64

	query := r.db.Model(&models.Organization{}).
		Preload("Account").
		Preload("Account.Industries").
		Preload("Technologies")

	// Filter by tenant (through Account)
	if tenantID != nil {
		query = query.Joins(`INNER JOIN "Account" ON "Account".id = "Organization".account_id`).
			Where(`"Account".tenant_id = ?`, *tenantID)
	}

	// Apply search
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(`LOWER("Organization".name) LIKE LOWER(?)`, searchTerm)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	query = ApplyPagination(query, params)

	if err := query.Find(&orgs).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[models.Organization]{
		Items:      orgs,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindByID returns an organization by ID
func (r *OrganizationRepository) FindByID(id int64) (*models.Organization, error) {
	var org models.Organization
	err := r.db.
		Preload("Account").
		Preload("Account.Industries").
		Preload("Technologies").
		Preload("FundingRounds").
		First(&org, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// FindByName searches organizations by name
func (r *OrganizationRepository) FindByName(name string, tenantID *int64, limit int) ([]models.Organization, error) {
	var orgs []models.Organization

	query := r.db.Model(&models.Organization{}).
		Preload("Account").
		Where(`LOWER("Organization".name) LIKE LOWER(?)`, "%"+name+"%")

	if tenantID != nil {
		query = query.Joins(`INNER JOIN "Account" ON "Account".id = "Organization".account_id`).
			Where(`"Account".tenant_id = ?`, *tenantID)
	}

	err := query.Limit(limit).Find(&orgs).Error
	return orgs, err
}

// FindOrCreateByName finds an organization by exact name or creates it with an Account
func (r *OrganizationRepository) FindOrCreateByName(name string, tenantID *int64, userID int64) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Preload("Account").
		Where(`LOWER("Organization".name) = LOWER(?)`, name).
		First(&org).Error

	if err == nil {
		return &org, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create Account first
	account := &models.Account{
		TenantID:  tenantID,
		Name:      name,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := r.db.Create(account).Error; err != nil {
		return nil, err
	}

	// Create Organization linked to Account
	newOrg := &models.Organization{
		AccountID: &account.ID,
		Name:      name,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := r.db.Create(newOrg).Error; err != nil {
		return nil, err
	}

	newOrg.Account = account
	return newOrg, nil
}

// Create creates a new organization
func (r *OrganizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}

// Update updates an organization
func (r *OrganizationRepository) Update(org *models.Organization, updates map[string]interface{}) error {
	return r.db.Model(&models.Organization{ID: org.ID}).Updates(updates).Error
}

// Delete deletes an organization by ID
func (r *OrganizationRepository) Delete(id int64) error {
	return r.db.Delete(&models.Organization{}, id).Error
}

// UpdateAccountIndustries replaces an account's industries
func (r *OrganizationRepository) UpdateAccountIndustries(accountID int64, industryIDs []int64) error {
	// Delete existing
	if err := r.db.Where("account_id = ?", accountID).Delete(&models.AccountIndustry{}).Error; err != nil {
		return err
	}
	// Insert new
	for _, indID := range industryIDs {
		ai := models.AccountIndustry{AccountID: accountID, IndustryID: indID}
		if err := r.db.Create(&ai).Error; err != nil {
			return err
		}
	}
	return nil
}

// OrgContactInput holds data for creating a contact for an org
type OrgContactInput struct {
	FirstName string
	LastName  string
	Email     *string
	Phone     *string
	Title     *string
	IsPrimary bool
}

// CreateContactForAccount creates a contact and links it to an account
func (r *OrganizationRepository) CreateContactForAccount(accountID int64, input OrgContactInput, userID int64) (*models.Contact, error) {
	contact := &models.Contact{
		FirstName: input.FirstName,
		LastName:  input.LastName,
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
