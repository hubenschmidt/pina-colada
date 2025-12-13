package repositories

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/shopspring/decimal"
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
		Preload("Technologies").
		Preload("EmployeeCountRange").
		Preload("FundingStage")

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

// GetContactsForAccount returns all contacts for an account
func (r *OrganizationRepository) GetContactsForAccount(accountID int64) ([]models.Contact, error) {
	var contacts []models.Contact
	err := r.db.Table(`"Contact"`).
		Joins(`INNER JOIN "Contact_Account" ON "Contact".id = "Contact_Account".contact_id`).
		Where(`"Contact_Account".account_id = ?`, accountID).
		Find(&contacts).Error
	return contacts, err
}

// UpdateContact updates a contact
func (r *OrganizationRepository) UpdateContact(contactID int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Contact{}).Where("id = ?", contactID).Updates(updates).Error
}

// DeleteContactFromAccount removes a contact's link to an account
func (r *OrganizationRepository) DeleteContactFromAccount(accountID int64, contactID int64) error {
	return r.db.Where("account_id = ? AND contact_id = ?", accountID, contactID).
		Delete(&models.ContactAccount{}).Error
}

// GetTechnologiesForOrg returns all technologies for an organization
func (r *OrganizationRepository) GetTechnologiesForOrg(orgID int64) ([]models.OrganizationTechnology, error) {
	var techs []models.OrganizationTechnology
	err := r.db.Where("organization_id = ?", orgID).Find(&techs).Error
	return techs, err
}

// AddTechnologyToOrg adds a technology to an organization
func (r *OrganizationRepository) AddTechnologyToOrg(orgID int64, techID int64, source *string, confidence *float64) error {
	tech := &models.OrganizationTechnology{
		OrganizationID: orgID,
		TechnologyID:   techID,
		Source:         source,
	}
	if confidence != nil {
		dec := decimal.NewFromFloat(*confidence)
		tech.Confidence = &dec
	}
	return r.db.Create(tech).Error
}

// RemoveTechnologyFromOrg removes a technology from an organization
func (r *OrganizationRepository) RemoveTechnologyFromOrg(orgID int64, techID int64) error {
	return r.db.Where("organization_id = ? AND technology_id = ?", orgID, techID).
		Delete(&models.OrganizationTechnology{}).Error
}

// GetFundingRoundsForOrg returns all funding rounds for an organization
func (r *OrganizationRepository) GetFundingRoundsForOrg(orgID int64) ([]models.FundingRound, error) {
	var rounds []models.FundingRound
	err := r.db.Where("organization_id = ?", orgID).Order("announced_date DESC").Find(&rounds).Error
	return rounds, err
}

// CreateFundingRound creates a funding round for an organization
func (r *OrganizationRepository) CreateFundingRound(round *models.FundingRound) error {
	return r.db.Create(round).Error
}

// DeleteFundingRound deletes a funding round
func (r *OrganizationRepository) DeleteFundingRound(roundID int64) error {
	return r.db.Delete(&models.FundingRound{}, roundID).Error
}

// GetSignalsForAccount returns signals for an account
func (r *OrganizationRepository) GetSignalsForAccount(accountID int64, signalType *string, limit int) ([]models.Signal, error) {
	var signals []models.Signal
	query := r.db.Where("account_id = ?", accountID)
	if signalType != nil {
		query = query.Where("signal_type = ?", *signalType)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&signals).Error
	return signals, err
}

// CreateSignal creates a signal for an account
func (r *OrganizationRepository) CreateSignal(signal *models.Signal) error {
	return r.db.Create(signal).Error
}

// DeleteSignal deletes a signal
func (r *OrganizationRepository) DeleteSignal(signalID int64) error {
	return r.db.Delete(&models.Signal{}, signalID).Error
}

// FundingRoundInput holds data for creating a funding round
type FundingRoundInput struct {
	OrganizationID int64
	RoundType      string
	Amount         *int64
	AnnouncedDate  *string
	LeadInvestor   *string
	SourceURL      *string
}

// CreateFundingRoundFromInput creates a funding round from input
func (r *OrganizationRepository) CreateFundingRoundFromInput(input *FundingRoundInput) (*models.FundingRound, error) {
	round := &models.FundingRound{
		OrganizationID: input.OrganizationID,
		RoundType:      input.RoundType,
		Amount:         input.Amount,
		LeadInvestor:   input.LeadInvestor,
		SourceURL:      input.SourceURL,
	}
	if input.AnnouncedDate != nil {
		if t, err := parseDate(*input.AnnouncedDate); err == nil {
			round.AnnouncedDate = &t
		}
	}
	if err := r.db.Create(round).Error; err != nil {
		return nil, err
	}
	return round, nil
}

// CreateSignalForAccount creates a signal for an account
func (r *OrganizationRepository) CreateSignalForAccount(accountID int64, input SignalCreateInput) (*models.Signal, error) {
	signal := &models.Signal{
		AccountID:      accountID,
		SignalType:     input.SignalType,
		Headline:       input.Headline,
		Description:    input.Description,
		Source:         input.Source,
		SourceURL:      input.SourceURL,
		Sentiment:      input.Sentiment,
		SignalDate:     input.SignalDate,
		RelevanceScore: input.RelevanceScore,
	}
	if err := r.db.Create(signal).Error; err != nil {
		return nil, err
	}
	return signal, nil
}

// CreateAccount creates an account
func (r *OrganizationRepository) CreateAccount(name string, tenantID *int64, userID int64) (*models.Account, error) {
	account := &models.Account{
		TenantID:  tenantID,
		Name:      name,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := r.db.Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

// OrganizationCreateInput holds data for creating an organization
type OrganizationCreateInput struct {
	AccountID            int64
	Name                 string
	Website              *string
	Phone                *string
	EmployeeCount        *int
	EmployeeCountRangeID *int64
	FundingStageID       *int64
	Description          *string
	RevenueRangeID       *int64
	FoundingYear         *int
	HeadquartersCity     *string
	HeadquartersState    *string
	HeadquartersCountry  *string
	CompanyType          *string
	LinkedInURL          *string
	CrunchbaseURL        *string
	UserID               int64
}

// CreateOrganization creates an organization
func (r *OrganizationRepository) CreateOrganization(input OrganizationCreateInput) (*models.Organization, error) {
	org := &models.Organization{
		AccountID:            &input.AccountID,
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
		CreatedBy:            input.UserID,
		UpdatedBy:            input.UserID,
	}
	if err := r.db.Create(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
