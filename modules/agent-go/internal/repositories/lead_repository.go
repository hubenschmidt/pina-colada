package repositories

import (
	"errors"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// OpportunityWithLead combines Opportunity with its Lead data
type OpportunityWithLead struct {
	models.Opportunity
	AccountName string `gorm:"-" json:"account_name"`
	AccountType string `gorm:"-" json:"account_type"`
	StatusName  string `gorm:"-" json:"status_name"`
}

// PartnershipWithLead combines Partnership with its Lead data
type PartnershipWithLead struct {
	models.Partnership
	AccountName string `gorm:"-" json:"account_name"`
	AccountType string `gorm:"-" json:"account_type"`
	StatusName  string `gorm:"-" json:"status_name"`
}

// LeadRepository handles lead-related data access
type LeadRepository struct {
	db *gorm.DB
}

// NewLeadRepository creates a new lead repository
func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// FindAllOpportunities returns paginated opportunities with lead data
func (r *LeadRepository) FindAllOpportunities(params PaginationParams, tenantID, projectID *int64) (*PaginatedResult[OpportunityWithLead], error) {
	var opportunities []models.Opportunity
	var totalCount int64

	query := r.db.Model(&models.Opportunity{}).
		Preload("Lead").
		Preload("Lead.CurrentStatus").
		Preload("Lead.Account").
		Preload("Lead.Account.Organizations").
		Preload("Lead.Account.Individuals")

	query = r.applyOpportunityFilters(query, tenantID, projectID, params.Search)

	countQuery := r.db.Model(&models.Opportunity{}).
		Joins("LEFT JOIN \"Lead\" ON \"Opportunity\".id = \"Lead\".id")
	countQuery = r.applyOpportunityFilters(countQuery, tenantID, projectID, params.Search)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&opportunities).Error; err != nil {
		return nil, err
	}

	results := r.mapOpportunitiesToDTO(opportunities)

	return &PaginatedResult[OpportunityWithLead]{
		Items:      results,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

func (r *LeadRepository) applyOpportunityFilters(query *gorm.DB, tenantID, projectID *int64, search string) *gorm.DB {
	query = query.Joins("LEFT JOIN \"Lead\" ON \"Opportunity\".id = \"Lead\".id")

	if tenantID != nil {
		query = query.Where("\"Lead\".tenant_id = ?", *tenantID)
	}

	if projectID != nil {
		query = query.Joins("INNER JOIN \"Lead_Project\" ON \"Lead_Project\".lead_id = \"Lead\".id").
			Where("\"Lead_Project\".project_id = ?", *projectID)
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("LOWER(\"Opportunity\".opportunity_name) LIKE LOWER(?)", searchTerm)
	}

	return query
}

func (r *LeadRepository) mapOpportunitiesToDTO(opportunities []models.Opportunity) []OpportunityWithLead {
	results := make([]OpportunityWithLead, len(opportunities))
	for i, opp := range opportunities {
		results[i] = OpportunityWithLead{
			Opportunity: opp,
			AccountName: resolveAccountNameFromLead(opp.Lead),
			AccountType: resolveAccountTypeFromLead(opp.Lead),
			StatusName:  resolveStatusName(opp.Lead, "Qualifying"),
		}
	}
	return results
}

// FindOpportunityByID returns an opportunity by ID with lead data
func (r *LeadRepository) FindOpportunityByID(id int64) (*OpportunityWithLead, error) {
	var opp models.Opportunity

	err := r.db.Model(&models.Opportunity{}).
		Preload("Lead").
		Preload("Lead.CurrentStatus").
		Preload("Lead.Account").
		Preload("Lead.Account.Organizations").
		Preload("Lead.Account.Individuals").
		Where("\"Opportunity\".id = ?", id).
		First(&opp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &OpportunityWithLead{
		Opportunity: opp,
		AccountName: resolveAccountNameFromLead(opp.Lead),
		AccountType: resolveAccountTypeFromLead(opp.Lead),
		StatusName:  resolveStatusName(opp.Lead, "Qualifying"),
	}, nil
}

// FindAllPartnerships returns paginated partnerships with lead data
func (r *LeadRepository) FindAllPartnerships(params PaginationParams, tenantID *int64) (*PaginatedResult[PartnershipWithLead], error) {
	var partnerships []models.Partnership
	var totalCount int64

	query := r.db.Model(&models.Partnership{}).
		Preload("Lead").
		Preload("Lead.CurrentStatus").
		Preload("Lead.Account").
		Preload("Lead.Account.Organizations").
		Preload("Lead.Account.Individuals")

	query = r.applyPartnershipFilters(query, tenantID, params.Search)

	countQuery := r.db.Model(&models.Partnership{}).
		Joins("LEFT JOIN \"Lead\" ON \"Partnership\".id = \"Lead\".id")
	countQuery = r.applyPartnershipFilters(countQuery, tenantID, params.Search)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&partnerships).Error; err != nil {
		return nil, err
	}

	results := r.mapPartnershipsToDTO(partnerships)

	return &PaginatedResult[PartnershipWithLead]{
		Items:      results,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

func (r *LeadRepository) applyPartnershipFilters(query *gorm.DB, tenantID *int64, search string) *gorm.DB {
	query = query.Joins("LEFT JOIN \"Lead\" ON \"Partnership\".id = \"Lead\".id")

	if tenantID != nil {
		query = query.Where("\"Lead\".tenant_id = ?", *tenantID)
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("LOWER(\"Partnership\".partnership_name) LIKE LOWER(?)", searchTerm)
	}

	return query
}

func (r *LeadRepository) mapPartnershipsToDTO(partnerships []models.Partnership) []PartnershipWithLead {
	results := make([]PartnershipWithLead, len(partnerships))
	for i, p := range partnerships {
		results[i] = PartnershipWithLead{
			Partnership: p,
			AccountName: resolveAccountNameFromLead(p.Lead),
			AccountType: resolveAccountTypeFromLead(p.Lead),
			StatusName:  resolveStatusName(p.Lead, "Exploring"),
		}
	}
	return results
}

// FindPartnershipByID returns a partnership by ID with lead data
func (r *LeadRepository) FindPartnershipByID(id int64) (*PartnershipWithLead, error) {
	var partnership models.Partnership

	err := r.db.Model(&models.Partnership{}).
		Preload("Lead").
		Preload("Lead.CurrentStatus").
		Preload("Lead.Account").
		Preload("Lead.Account.Organizations").
		Preload("Lead.Account.Individuals").
		Where("\"Partnership\".id = ?", id).
		First(&partnership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &PartnershipWithLead{
		Partnership: partnership,
		AccountName: resolveAccountNameFromLead(partnership.Lead),
		AccountType: resolveAccountTypeFromLead(partnership.Lead),
		StatusName:  resolveStatusName(partnership.Lead, "Exploring"),
	}, nil
}

func resolveAccountNameFromLead(lead *models.Lead) string {
	if lead == nil || lead.Account == nil {
		return ""
	}

	if len(lead.Account.Organizations) > 0 {
		return lead.Account.Organizations[0].Name
	}

	if len(lead.Account.Individuals) > 0 {
		ind := lead.Account.Individuals[0]
		return formatIndividualName(&ind.FirstName, &ind.LastName)
	}

	return ""
}

func resolveAccountTypeFromLead(lead *models.Lead) string {
	if lead == nil || lead.Account == nil {
		return "Organization"
	}

	if len(lead.Account.Organizations) > 0 {
		return "Organization"
	}

	if len(lead.Account.Individuals) > 0 {
		return "Individual"
	}

	return "Organization"
}

func resolveStatusName(lead *models.Lead, defaultStatus string) string {
	if lead == nil || lead.CurrentStatus == nil {
		return defaultStatus
	}
	return lead.CurrentStatus.Name
}

func formatIndividualName(firstName *string, lastName *string) string {
	fn := derefString(firstName)
	ln := derefString(lastName)
	if ln != "" && fn != "" {
		return ln + ", " + fn
	}
	if ln != "" {
		return ln
	}
	return fn
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
