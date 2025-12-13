package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// OpportunityWithLead combines Opportunity with its Lead data
type OpportunityWithLead struct {
	models.Opportunity
	Lead             *models.Lead `gorm:"foreignKey:ID"`
	AccountName      string       `gorm:"-"`
	AccountType      string       `gorm:"-"`
	StatusName       string       `gorm:"-"`
	OrganizationName *string      `gorm:"column:organization_name"`
	FirstName        *string      `gorm:"column:first_name"`
	LastName         *string      `gorm:"column:last_name"`
	CurrentStatus    *string      `gorm:"column:status_name"`
}

// PartnershipWithLead combines Partnership with its Lead data
type PartnershipWithLead struct {
	models.Partnership
	Lead             *models.Lead `gorm:"foreignKey:ID"`
	AccountName      string       `gorm:"-"`
	AccountType      string       `gorm:"-"`
	StatusName       string       `gorm:"-"`
	OrganizationName *string      `gorm:"column:organization_name"`
	FirstName        *string      `gorm:"column:first_name"`
	LastName         *string      `gorm:"column:last_name"`
	CurrentStatus    *string      `gorm:"column:status_name"`
}

// LeadRepository handles lead-related data access
type LeadRepository struct {
	db *gorm.DB
}

// NewLeadRepository creates a new lead repository
func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// resolveAccountName extracts account name and type from org/individual data
func resolveAccountName(orgName *string, firstName *string, lastName *string) (string, string) {
	if orgName != nil && *orgName != "" {
		return *orgName, "Organization"
	}
	if firstName != nil || lastName != nil {
		fn := ""
		ln := ""
		if firstName != nil {
			fn = *firstName
		}
		if lastName != nil {
			ln = *lastName
		}
		if ln != "" && fn != "" {
			return ln + ", " + fn, "Individual"
		} else if ln != "" {
			return ln, "Individual"
		} else if fn != "" {
			return fn, "Individual"
		}
	}
	return "", "Organization"
}

// FindAllOpportunities returns paginated opportunities with lead data
func (r *LeadRepository) FindAllOpportunities(params PaginationParams, tenantID, projectID *int64) (*PaginatedResult[OpportunityWithLead], error) {
	var opportunities []OpportunityWithLead
	var totalCount int64

	query := r.db.Table(`"Opportunity"`).
		Select(`"Opportunity".*, "Lead".current_status_id, "Lead".account_id, "Lead".tenant_id, "Organization".name as organization_name, "Individual".first_name, "Individual".last_name, "Status".name as status_name`).
		Joins(`LEFT JOIN "Lead" ON "Opportunity".id = "Lead".id`).
		Joins(`LEFT JOIN "Account" ON "Account".id = "Lead".account_id`).
		Joins(`LEFT JOIN "Organization" ON "Organization".account_id = "Account".id`).
		Joins(`LEFT JOIN "Individual" ON "Individual".account_id = "Account".id`).
		Joins(`LEFT JOIN "Status" ON "Status".id = "Lead".current_status_id`)

	if tenantID != nil {
		query = query.Where(`"Lead".tenant_id = ?`, *tenantID)
	}

	if projectID != nil {
		query = query.Joins(`INNER JOIN "Lead_Project" ON "Lead_Project".lead_id = "Lead".id`).
			Where(`"Lead_Project".project_id = ?`, *projectID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(`LOWER("Opportunity".opportunity_name) LIKE LOWER(?)`, searchTerm)
	}

	// Count before pagination
	countQuery := r.db.Table(`"Opportunity"`).
		Joins(`LEFT JOIN "Lead" ON "Opportunity".id = "Lead".id`)
	if tenantID != nil {
		countQuery = countQuery.Where(`"Lead".tenant_id = ?`, *tenantID)
	}
	if projectID != nil {
		countQuery = countQuery.Joins(`INNER JOIN "Lead_Project" ON "Lead_Project".lead_id = "Lead".id`).
			Where(`"Lead_Project".project_id = ?`, *projectID)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		countQuery = countQuery.Where(`LOWER("Opportunity".opportunity_name) LIKE LOWER(?)`, searchTerm)
	}
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&opportunities).Error; err != nil {
		return nil, err
	}

	// Resolve account name and type
	for i := range opportunities {
		opportunities[i].AccountName, opportunities[i].AccountType = resolveAccountName(
			opportunities[i].OrganizationName,
			opportunities[i].FirstName,
			opportunities[i].LastName,
		)
		if opportunities[i].CurrentStatus != nil {
			opportunities[i].StatusName = *opportunities[i].CurrentStatus
		} else {
			opportunities[i].StatusName = "Qualifying"
		}
	}

	return &PaginatedResult[OpportunityWithLead]{
		Items:      opportunities,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindOpportunityByID returns an opportunity by ID with lead data
func (r *LeadRepository) FindOpportunityByID(id int64) (*OpportunityWithLead, error) {
	var opp OpportunityWithLead

	err := r.db.Table(`"Opportunity"`).
		Select(`"Opportunity".*, "Lead".current_status_id, "Lead".account_id, "Lead".tenant_id`).
		Joins(`LEFT JOIN "Lead" ON "Opportunity".id = "Lead".id`).
		Where(`"Opportunity".id = ?`, id).
		First(&opp).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &opp, nil
}

// FindAllPartnerships returns paginated partnerships with lead data
func (r *LeadRepository) FindAllPartnerships(params PaginationParams, tenantID *int64) (*PaginatedResult[PartnershipWithLead], error) {
	var partnerships []PartnershipWithLead
	var totalCount int64

	query := r.db.Table(`"Partnership"`).
		Select(`"Partnership".*, "Lead".current_status_id, "Lead".account_id, "Lead".tenant_id, "Organization".name as organization_name, "Individual".first_name, "Individual".last_name, "Status".name as status_name`).
		Joins(`LEFT JOIN "Lead" ON "Partnership".id = "Lead".id`).
		Joins(`LEFT JOIN "Account" ON "Account".id = "Lead".account_id`).
		Joins(`LEFT JOIN "Organization" ON "Organization".account_id = "Account".id`).
		Joins(`LEFT JOIN "Individual" ON "Individual".account_id = "Account".id`).
		Joins(`LEFT JOIN "Status" ON "Status".id = "Lead".current_status_id`)

	if tenantID != nil {
		query = query.Where(`"Lead".tenant_id = ?`, *tenantID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(`LOWER("Partnership".partnership_name) LIKE LOWER(?)`, searchTerm)
	}

	// Count before pagination
	countQuery := r.db.Table(`"Partnership"`).
		Joins(`LEFT JOIN "Lead" ON "Partnership".id = "Lead".id`)
	if tenantID != nil {
		countQuery = countQuery.Where(`"Lead".tenant_id = ?`, *tenantID)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		countQuery = countQuery.Where(`LOWER("Partnership".partnership_name) LIKE LOWER(?)`, searchTerm)
	}
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&partnerships).Error; err != nil {
		return nil, err
	}

	// Resolve account name and type
	for i := range partnerships {
		partnerships[i].AccountName, partnerships[i].AccountType = resolveAccountName(
			partnerships[i].OrganizationName,
			partnerships[i].FirstName,
			partnerships[i].LastName,
		)
		if partnerships[i].CurrentStatus != nil {
			partnerships[i].StatusName = *partnerships[i].CurrentStatus
		} else {
			partnerships[i].StatusName = "Exploring"
		}
	}

	return &PaginatedResult[PartnershipWithLead]{
		Items:      partnerships,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindPartnershipByID returns a partnership by ID with lead data
func (r *LeadRepository) FindPartnershipByID(id int64) (*PartnershipWithLead, error) {
	var partnership PartnershipWithLead

	err := r.db.Table(`"Partnership"`).
		Select(`"Partnership".*, "Lead".current_status_id, "Lead".account_id, "Lead".tenant_id`).
		Joins(`LEFT JOIN "Lead" ON "Partnership".id = "Lead".id`).
		Where(`"Partnership".id = ?`, id).
		First(&partnership).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &partnership, nil
}
