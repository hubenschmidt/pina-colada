package repositories

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	apperrors "agent/internal/errors"
	"agent/internal/models"
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

	countQuery := r.db.Model(&models.Opportunity{})
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
		leadIDs := r.db.Model(&models.LeadProject{}).Select("lead_id").Where("project_id = ?", *projectID)
		query = query.Where("\"Lead\".id IN (?)", leadIDs)
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

	countQuery := r.db.Model(&models.Partnership{})
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

func decimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseDatePtr(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := parseDate(*s)
	if err != nil {
		return nil
	}
	return &t
}

func decimalPtrFromFloat(f *float64) *decimal.Decimal {
	if f == nil {
		return nil
	}
	d := decimal.NewFromFloat(*f)
	return &d
}

func int16PtrFromInt(i *int) *int16 {
	if i == nil {
		return nil
	}
	v := int16(*i)
	return &v
}

func createLeadProjects(tx *gorm.DB, leadID int64, projectIDs []int64) error {
	for _, pid := range projectIDs {
		lp := models.LeadProject{LeadID: leadID, ProjectID: pid}
		if err := tx.Create(&lp).Error; err != nil {
			return err
		}
	}
	return nil
}

// OpportunityCreateInput contains data needed to create an opportunity
type OpportunityCreateInput struct {
	TenantID          *int64
	UserID            int64
	AccountID         *int64
	CurrentStatusID   *int64
	Source            string
	OpportunityName   string
	EstimatedValue    *float64
	Probability       *int
	ExpectedCloseDate *string
	Description       *string
	ProjectIDs        []int64
}

// CreateOpportunity creates a new opportunity with its Lead and Deal
func (r *LeadRepository) CreateOpportunity(input OpportunityCreateInput) (int64, error) {
	var oppID int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		deal := &models.Deal{
			TenantID:  input.TenantID,
			Name:      input.OpportunityName,
			CreatedBy: input.UserID,
			UpdatedBy: input.UserID,
		}
		if err := tx.Create(deal).Error; err != nil {
			return err
		}

		lead := &models.Lead{
			TenantID:        input.TenantID,
			DealID:          deal.ID,
			Type:            "Opportunity",
			Source:          &input.Source,
			AccountID:       input.AccountID,
			CurrentStatusID: input.CurrentStatusID,
			CreatedBy:       input.UserID,
			UpdatedBy:       input.UserID,
		}
		if err := tx.Create(lead).Error; err != nil {
			return err
		}

		opp := &models.Opportunity{
			ID:                lead.ID,
			OpportunityName:   input.OpportunityName,
			Description:       input.Description,
			EstimatedValue:    decimalPtrFromFloat(input.EstimatedValue),
			Probability:       int16PtrFromInt(input.Probability),
			ExpectedCloseDate: parseDatePtr(input.ExpectedCloseDate),
		}
		if err := tx.Create(opp).Error; err != nil {
			return err
		}

		oppID = opp.ID
		return createLeadProjects(tx, lead.ID, input.ProjectIDs)
	})
	return oppID, err
}

// OpportunityUpdateInput contains data for updating an opportunity
type OpportunityUpdateInput struct {
	UserID            int64
	OpportunityName   *string
	EstimatedValue    *float64
	Probability       *int
	ExpectedCloseDate *string
	Description       *string
	CurrentStatusID   *int64
}

// UpdateOpportunity updates an existing opportunity
func (r *LeadRepository) UpdateOpportunity(id int64, input OpportunityUpdateInput) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		updates := buildOpportunityUpdates(input)
		if len(updates) > 0 {
			err := tx.Model(&models.Opportunity{}).Where("id = ?", id).Updates(updates).Error
			if err != nil {
				return err
			}
		}
		return updateLeadStatus(tx, id, input.CurrentStatusID, input.UserID)
	})
}

func buildOpportunityUpdates(input OpportunityUpdateInput) map[string]interface{} {
	updates := map[string]interface{}{}
	if input.OpportunityName != nil {
		updates["opportunity_name"] = *input.OpportunityName
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.EstimatedValue != nil {
		updates["estimated_value"] = decimalFromFloat(*input.EstimatedValue)
	}
	if input.Probability != nil {
		updates["probability"] = int16(*input.Probability)
	}
	if input.ExpectedCloseDate != nil {
		updates["expected_close_date"] = parseDatePtr(input.ExpectedCloseDate)
	}
	return updates
}

func updateLeadStatus(tx *gorm.DB, id int64, statusID *int64, userID int64) error {
	if statusID == nil {
		return nil
	}
	return tx.Model(&models.Lead{}).Where("id = ?", id).Updates(map[string]interface{}{
		"current_status_id": *statusID,
		"updated_by":        userID,
	}).Error
}

// DeleteOpportunity deletes an opportunity and its Lead/Deal
func (r *LeadRepository) DeleteOpportunity(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var lead models.Lead
		if err := tx.First(&lead, id).Error; err != nil {
			return err
		}

		if err := tx.Where("lead_id = ?", id).Delete(&models.LeadProject{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Opportunity{}, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Lead{}, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Deal{}, lead.DealID).Error; err != nil {
			return err
		}
		return nil
	})
}

// PartnershipCreateInput contains data needed to create a partnership
type PartnershipCreateInput struct {
	TenantID        *int64
	UserID          int64
	AccountID       *int64
	CurrentStatusID *int64
	Source          string
	PartnershipName string
	PartnershipType *string
	StartDate       *string
	EndDate         *string
	Description     *string
	ProjectIDs      []int64
}

// CreatePartnership creates a new partnership with its Lead and Deal
func (r *LeadRepository) CreatePartnership(input PartnershipCreateInput) (int64, error) {
	var partnershipID int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		deal := &models.Deal{
			TenantID:  input.TenantID,
			Name:      input.PartnershipName,
			CreatedBy: input.UserID,
			UpdatedBy: input.UserID,
		}
		if err := tx.Create(deal).Error; err != nil {
			return err
		}

		lead := &models.Lead{
			TenantID:        input.TenantID,
			DealID:          deal.ID,
			Type:            "Partnership",
			Source:          &input.Source,
			AccountID:       input.AccountID,
			CurrentStatusID: input.CurrentStatusID,
			CreatedBy:       input.UserID,
			UpdatedBy:       input.UserID,
		}
		if err := tx.Create(lead).Error; err != nil {
			return err
		}

		partnership := &models.Partnership{
			ID:              lead.ID,
			PartnershipName: input.PartnershipName,
			PartnershipType: input.PartnershipType,
			Description:     input.Description,
			StartDate:       parseDatePtr(input.StartDate),
			EndDate:         parseDatePtr(input.EndDate),
		}
		if err := tx.Create(partnership).Error; err != nil {
			return err
		}

		partnershipID = partnership.ID
		return createLeadProjects(tx, lead.ID, input.ProjectIDs)
	})
	return partnershipID, err
}

// PartnershipUpdateInput contains data for updating a partnership
type PartnershipUpdateInput struct {
	UserID          int64
	PartnershipName *string
	PartnershipType *string
	StartDate       *string
	EndDate         *string
	Description     *string
	CurrentStatusID *int64
}

// UpdatePartnership updates an existing partnership
func (r *LeadRepository) UpdatePartnership(id int64, input PartnershipUpdateInput) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		updates := buildPartnershipUpdates(input)
		if len(updates) > 0 {
			err := tx.Model(&models.Partnership{}).Where("id = ?", id).Updates(updates).Error
			if err != nil {
				return err
			}
		}
		return updateLeadStatus(tx, id, input.CurrentStatusID, input.UserID)
	})
}

func buildPartnershipUpdates(input PartnershipUpdateInput) map[string]interface{} {
	updates := map[string]interface{}{}
	if input.PartnershipName != nil {
		updates["partnership_name"] = *input.PartnershipName
	}
	if input.PartnershipType != nil {
		updates["partnership_type"] = *input.PartnershipType
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.StartDate != nil {
		updates["start_date"] = parseDatePtr(input.StartDate)
	}
	if input.EndDate != nil {
		updates["end_date"] = parseDatePtr(input.EndDate)
	}
	return updates
}

// DeletePartnership deletes a partnership and its Lead/Deal
func (r *LeadRepository) DeletePartnership(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var lead models.Lead
		if err := tx.First(&lead, id).Error; err != nil {
			return err
		}

		if err := tx.Where("lead_id = ?", id).Delete(&models.LeadProject{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Partnership{}, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Lead{}, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Deal{}, lead.DealID).Error; err != nil {
			return err
		}
		return nil
	})
}
