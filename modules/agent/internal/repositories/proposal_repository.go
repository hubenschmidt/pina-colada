package repositories

import (
	"encoding/json"
	"errors"
	"time"

	apperrors "agent/internal/errors"
	"agent/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Proposal status constants
const (
	ProposalStatusPending  = "pending"
	ProposalStatusApproved = "approved"
	ProposalStatusRejected = "rejected"
	ProposalStatusExecuted = "executed"
	ProposalStatusFailed   = "failed"
)

// Operation constants
const (
	OperationCreate = "create"
	OperationUpdate = "update"
	OperationDelete = "delete"
)

// SupportedEntities lists entity types that support proposals
var SupportedEntities = []string{
	"account",
	"account_industry",
	"account_project",
	"account_relationship",
	"activity",
	"agent_config_user_selection",
	"agent_metric",
	"agent_node_config",
	"asset",
	"comment",
	"contact",
	"contact_account",
	"conversation",
	"conversation_message",
	"deal",
	"document",
	"entity_asset",
	"entity_tag",
	"funding_round",
	"individual",
	"individual_relationship",
	"industry",
	"job",
	"lead",
	"lead_project",
	"note",
	"opportunity",
	"organization",
	"organization_relationship",
	"organization_technology",
	"partnership",
	"project",
	"provenance",
	"saved_report",
	"saved_report_project",
	"signal",
	"status",
	"tag",
	"task",
	"technology",
	"usage_analytics",
}

// ProposalDTO represents proposal data returned from repository
type ProposalDTO struct {
	ID                   int64
	TenantID             int64
	ProposedByID         int64
	EntityType           string
	EntityID             *int64
	Operation            string
	Payload              datatypes.JSON
	Status               string
	ValidationErrors     datatypes.JSON
	ReviewedByID         *int64
	ReviewedAt           *time.Time
	ExecutedAt           *time.Time
	ErrorMessage         *string
	Source               *string
	AutomationConfigID   *int64
	AutomationConfigName *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ProposalCreateInput contains data needed to create a proposal
type ProposalCreateInput struct {
	TenantID           int64
	ProposedByID       int64
	EntityType         string
	EntityID           *int64
	Operation          string
	Payload            datatypes.JSON
	ValidationErrors   datatypes.JSON
	Source             *string
	AutomationConfigID *int64
}

// ProposalRepository handles proposal data access
type ProposalRepository struct {
	db *gorm.DB
}

// NewProposalRepository creates a new proposal repository
func NewProposalRepository(db *gorm.DB) *ProposalRepository {
	return &ProposalRepository{db: db}
}

// Create creates a new proposal
func (r *ProposalRepository) Create(input ProposalCreateInput) (int64, error) {
	proposal := &models.AgentProposal{
		TenantID:           input.TenantID,
		ProposedByID:       input.ProposedByID,
		EntityType:         input.EntityType,
		EntityID:           input.EntityID,
		Operation:          input.Operation,
		Payload:            input.Payload,
		ValidationErrors:   input.ValidationErrors,
		Status:             ProposalStatusPending,
		Source:             input.Source,
		AutomationConfigID: input.AutomationConfigID,
	}
	if err := r.db.Create(proposal).Error; err != nil {
		return 0, err
	}
	return proposal.ID, nil
}

// FindByID returns a proposal by ID
func (r *ProposalRepository) FindByID(id int64) (*ProposalDTO, error) {
	var proposal models.AgentProposal
	err := r.db.First(&proposal, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return modelToDTO(&proposal), nil
}

// GetAllPendingIDs returns all pending proposal IDs for a tenant
func (r *ProposalRepository) GetAllPendingIDs(tenantID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Model(&models.AgentProposal{}).
		Where("tenant_id = ? AND status = ?", tenantID, ProposalStatusPending).
		Pluck("id", &ids).Error
	return ids, err
}

// CountPendingByConfigID returns count of pending proposals for a specific automation config
func (r *ProposalRepository) CountPendingByConfigID(configID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.AgentProposal{}).
		Where("automation_config_id = ? AND status = ?", configID, ProposalStatusPending).
		Count(&count).Error
	return count, err
}

// ProposalFilterParams contains optional filters for proposal queries
type ProposalFilterParams struct {
	AutomationConfigID *int64
}

// FindPending returns pending proposals for a tenant with pagination
func (r *ProposalRepository) FindPending(tenantID int64, params PaginationParams, filters *ProposalFilterParams) (*PaginatedResult[ProposalDTO], error) {
	var totalCount int64

	// Build base query for count
	countQuery := r.db.Model(&models.AgentProposal{}).
		Where("\"Agent_Proposal\".tenant_id = ? AND \"Agent_Proposal\".status = ?", tenantID, ProposalStatusPending)

	if filters != nil && filters.AutomationConfigID != nil {
		countQuery = countQuery.Where("\"Agent_Proposal\".automation_config_id = ?", *filters.AutomationConfigID)
	}

	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Build query with join for config name
	type proposalWithConfig struct {
		models.AgentProposal
		AutomationConfigName *string `gorm:"column:automation_config_name"`
	}

	var results []proposalWithConfig

	query := r.db.Table("\"Agent_Proposal\"").
		Select("\"Agent_Proposal\".*, \"Automation_Config\".name as automation_config_name").
		Joins("LEFT JOIN \"Automation_Config\" ON \"Agent_Proposal\".automation_config_id = \"Automation_Config\".id").
		Where("\"Agent_Proposal\".tenant_id = ? AND \"Agent_Proposal\".status = ?", tenantID, ProposalStatusPending)

	if filters != nil && filters.AutomationConfigID != nil {
		query = query.Where("\"Agent_Proposal\".automation_config_id = ?", *filters.AutomationConfigID)
	}

	query = ApplyPagination(query, params)
	query = query.Order("\"Agent_Proposal\".created_at DESC")

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	dtos := make([]ProposalDTO, len(results))
	for i := range results {
		dto := modelToDTO(&results[i].AgentProposal)
		dto.AutomationConfigName = results[i].AutomationConfigName
		dtos[i] = *dto
	}

	return &PaginatedResult[ProposalDTO]{
		Items:      dtos,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// UpdateStatus updates a proposal's status
func (r *ProposalRepository) UpdateStatus(id int64, status string, reviewerID *int64) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if reviewerID != nil {
		updates["reviewed_by_id"] = *reviewerID
		updates["reviewed_at"] = time.Now()
	}
	return r.db.Model(&models.AgentProposal{}).Where("id = ?", id).Updates(updates).Error
}

// MarkExecuted marks a proposal as executed
func (r *ProposalRepository) MarkExecuted(id int64) error {
	return r.db.Model(&models.AgentProposal{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      ProposalStatusExecuted,
			"executed_at": time.Now(),
			"updated_at":  time.Now(),
		}).Error
}

// MarkFailed marks a proposal as failed with error message
func (r *ProposalRepository) MarkFailed(id int64, errMsg string) error {
	return r.db.Model(&models.AgentProposal{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        ProposalStatusFailed,
			"error_message": errMsg,
			"updated_at":    time.Now(),
		}).Error
}

// SetError sets error message without changing status (keeps proposal visible for retry)
func (r *ProposalRepository) SetError(id int64, errMsg string) error {
	return r.db.Model(&models.AgentProposal{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"error_message": errMsg,
			"updated_at":    time.Now(),
		}).Error
}

// FindByIDs returns proposals by IDs
func (r *ProposalRepository) FindByIDs(ids []int64) ([]ProposalDTO, error) {
	var proposals []models.AgentProposal
	if err := r.db.Where("id IN ?", ids).Find(&proposals).Error; err != nil {
		return nil, err
	}
	dtos := make([]ProposalDTO, len(proposals))
	for i := range proposals {
		dtos[i] = *modelToDTO(&proposals[i])
	}
	return dtos, nil
}

// UpdatePayload updates a proposal's payload and validation errors
func (r *ProposalRepository) UpdatePayload(id int64, payload, validationErrors datatypes.JSON) error {
	return r.db.Model(&models.AgentProposal{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"payload":           payload,
			"validation_errors": validationErrors,
			"updated_at":        time.Now(),
		}).Error
}

func modelToDTO(p *models.AgentProposal) *ProposalDTO {
	return &ProposalDTO{
		ID:                 p.ID,
		TenantID:           p.TenantID,
		ProposedByID:       p.ProposedByID,
		EntityType:         p.EntityType,
		EntityID:           p.EntityID,
		Operation:          p.Operation,
		Payload:            p.Payload,
		Status:             p.Status,
		ValidationErrors:   p.ValidationErrors,
		ReviewedByID:       p.ReviewedByID,
		ReviewedAt:         p.ReviewedAt,
		ExecutedAt:         p.ExecutedAt,
		ErrorMessage:       p.ErrorMessage,
		Source:             p.Source,
		AutomationConfigID: p.AutomationConfigID,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

// AutomationProposalDTO represents a simplified proposal for automation digest
type AutomationProposalDTO struct {
	ID        int64
	Status    string
	CreatedAt time.Time
	JobTitle  string
	Account   string
}

// GetAutomationProposals returns proposals created by automation in the given time range
func (r *ProposalRepository) GetAutomationProposals(tenantID, userID int64, since time.Time) ([]AutomationProposalDTO, error) {
	var proposals []models.AgentProposal
	err := r.db.Where("tenant_id = ? AND proposed_by_id = ? AND source = ? AND created_at >= ?",
		tenantID, userID, "automation", since).
		Order("created_at DESC").
		Find(&proposals).Error
	if err != nil {
		return nil, err
	}
	return r.proposalsToDTO(proposals), nil
}

// GetProposalsByConfigID returns proposals for a specific automation config
func (r *ProposalRepository) GetProposalsByConfigID(configID int64) ([]AutomationProposalDTO, error) {
	var proposals []models.AgentProposal
	err := r.db.Where("automation_config_id = ?", configID).
		Order("created_at DESC").
		Find(&proposals).Error
	if err != nil {
		return nil, err
	}
	return r.proposalsToDTO(proposals), nil
}

func (r *ProposalRepository) proposalsToDTO(proposals []models.AgentProposal) []AutomationProposalDTO {
	result := make([]AutomationProposalDTO, len(proposals))
	for i := range proposals {
		p := &proposals[i]
		dto := AutomationProposalDTO{
			ID:        p.ID,
			Status:    p.Status,
			CreatedAt: p.CreatedAt,
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(p.Payload, &payload); err == nil {
			if title, ok := payload["job_title"].(string); ok {
				dto.JobTitle = title
			}
			if account, ok := payload["account"].(string); ok {
				dto.Account = account
			}
		}
		result[i] = dto
	}
	return result
}

// PendingJobProposal holds minimal proposal data for deduplication
type PendingJobProposal struct {
	URL     string
	Title   string
	Company string
}

// GetPendingJobProposals returns pending job proposals with URL, title, and company for deduplication
func (r *ProposalRepository) GetPendingJobProposals(tenantID int64) ([]PendingJobProposal, error) {
	var proposals []models.AgentProposal
	err := r.db.Where("tenant_id = ? AND status = ? AND entity_type = ?", tenantID, ProposalStatusPending, "job").
		Find(&proposals).Error
	if err != nil {
		return nil, err
	}

	result := make([]PendingJobProposal, 0, len(proposals))
	for _, p := range proposals {
		var payload map[string]interface{}
		if err := json.Unmarshal(p.Payload, &payload); err != nil {
			continue
		}
		prop := PendingJobProposal{}
		if url, ok := payload["url"].(string); ok {
			prop.URL = url
		}
		if title, ok := payload["job_title"].(string); ok {
			prop.Title = title
		}
		if account, ok := payload["account"].(string); ok {
			prop.Company = account
		}
		result = append(result, prop)
	}
	return result, nil
}

// GetRejectedJobProposals returns user-rejected job proposals for deduplication
func (r *ProposalRepository) GetRejectedJobProposals(tenantID int64) ([]PendingJobProposal, error) {
	var proposals []models.AgentProposal
	err := r.db.Where("tenant_id = ? AND status = ? AND entity_type = ?", tenantID, ProposalStatusRejected, "job").
		Find(&proposals).Error
	if err != nil {
		return nil, err
	}

	result := make([]PendingJobProposal, 0, len(proposals))
	for _, p := range proposals {
		var payload map[string]interface{}
		if err := json.Unmarshal(p.Payload, &payload); err != nil {
			continue
		}
		prop := PendingJobProposal{}
		if url, ok := payload["url"].(string); ok {
			prop.URL = url
		}
		if title, ok := payload["job_title"].(string); ok {
			prop.Title = title
		}
		if account, ok := payload["account"].(string); ok {
			prop.Company = account
		}
		result = append(result, prop)
	}
	return result, nil
}
