package repositories

import (
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
	ID               int64
	TenantID         int64
	ProposedByID     int64
	EntityType       string
	EntityID         *int64
	Operation        string
	Payload          datatypes.JSON
	Status           string
	ValidationErrors datatypes.JSON
	ReviewedByID     *int64
	ReviewedAt       *time.Time
	ExecutedAt       *time.Time
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ProposalCreateInput contains data needed to create a proposal
type ProposalCreateInput struct {
	TenantID         int64
	ProposedByID     int64
	EntityType       string
	EntityID         *int64
	Operation        string
	Payload          datatypes.JSON
	ValidationErrors datatypes.JSON
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
		TenantID:         input.TenantID,
		ProposedByID:     input.ProposedByID,
		EntityType:       input.EntityType,
		EntityID:         input.EntityID,
		Operation:        input.Operation,
		Payload:          input.Payload,
		ValidationErrors: input.ValidationErrors,
		Status:           ProposalStatusPending,
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

// FindPending returns pending proposals for a tenant with pagination
func (r *ProposalRepository) FindPending(tenantID int64, params PaginationParams) (*PaginatedResult[ProposalDTO], error) {
	var proposals []models.AgentProposal
	var totalCount int64

	query := r.db.Model(&models.AgentProposal{}).
		Where("tenant_id = ? AND status = ?", tenantID, ProposalStatusPending)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)
	query = query.Order("created_at DESC")

	if err := query.Find(&proposals).Error; err != nil {
		return nil, err
	}

	dtos := make([]ProposalDTO, len(proposals))
	for i := range proposals {
		dtos[i] = *modelToDTO(&proposals[i])
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
		ID:               p.ID,
		TenantID:         p.TenantID,
		ProposedByID:     p.ProposedByID,
		EntityType:       p.EntityType,
		EntityID:         p.EntityID,
		Operation:        p.Operation,
		Payload:          p.Payload,
		Status:           p.Status,
		ValidationErrors: p.ValidationErrors,
		ReviewedByID:     p.ReviewedByID,
		ReviewedAt:       p.ReviewedAt,
		ExecutedAt:       p.ExecutedAt,
		ErrorMessage:     p.ErrorMessage,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}
