package services

import (
	"encoding/json"
	"errors"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/validation"
	"gorm.io/datatypes"
)

var (
	ErrProposalNotFound           = errors.New("proposal not found")
	ErrProposalNotPending         = errors.New("proposal is not in pending status")
	ErrProposalInvalidEntity      = errors.New("invalid entity type")
	ErrProposalInvalidOperation   = errors.New("invalid operation")
	ErrProposalExecutionFailed    = errors.New("proposal execution failed")
	ErrProposalHasValidationErrors = errors.New("proposal has validation errors - fix payload before approving")
)

// ProposeResult is a simplified result for tool callers
type ProposeResult struct {
	ProposalID int64
	Status     string
	Message    string
}

// ProposalExecutor executes approved proposals
type ProposalExecutor interface {
	Execute(entityType, operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error
}

// ProposalService handles proposal business logic
type ProposalService struct {
	proposalRepo   *repositories.ProposalRepository
	approvalRepo   *repositories.ApprovalConfigRepository
	executor       ProposalExecutor
}

// NewProposalService creates a new proposal service
func NewProposalService(
	proposalRepo *repositories.ProposalRepository,
	approvalRepo *repositories.ApprovalConfigRepository,
	executor ProposalExecutor,
) *ProposalService {
	return &ProposalService{
		proposalRepo:   proposalRepo,
		approvalRepo:   approvalRepo,
		executor:       executor,
	}
}

// ProposeOperation creates a proposal or executes immediately if approval not required
func (s *ProposalService) ProposeOperation(
	tenantID, userID int64,
	entityType string,
	entityID *int64,
	operation string,
	payload datatypes.JSON,
) (*serializers.ProposalResponse, error) {
	if !isValidEntityType(entityType) {
		return nil, ErrProposalInvalidEntity
	}
	if !isValidOperation(operation) {
		return nil, ErrProposalInvalidOperation
	}

	requiresApproval, err := s.approvalRepo.RequiresApproval(tenantID, entityType)
	if err != nil {
		return nil, err
	}

	// Execute immediately if approval not required
	if !requiresApproval {
		if err := s.executor.Execute(entityType, operation, entityID, payload, tenantID, userID); err != nil {
			return nil, err
		}
		return &serializers.ProposalResponse{
			Status:  repositories.ProposalStatusExecuted,
			Message: "operation executed immediately (approval not required)",
		}, nil
	}

	// Validate payload and store any errors
	validationErrors := validation.ValidatePayload(entityType, operation, payload)
	var validationErrorsJSON datatypes.JSON
	if len(validationErrors) > 0 {
		validationErrorsJSON, _ = json.Marshal(validationErrors)
	}

	// Queue for approval
	input := repositories.ProposalCreateInput{
		TenantID:         tenantID,
		ProposedByID:     userID,
		EntityType:       entityType,
		EntityID:         entityID,
		Operation:        operation,
		Payload:          payload,
		ValidationErrors: validationErrorsJSON,
	}

	proposalID, err := s.proposalRepo.Create(input)
	if err != nil {
		return nil, err
	}

	proposal, err := s.proposalRepo.FindByID(proposalID)
	if err != nil {
		return nil, err
	}

	return proposalToResponse(proposal), nil
}

// ProposeOperationBytes is a convenience method that takes []byte payload (for tool callers)
func (s *ProposalService) ProposeOperationBytes(
	tenantID, userID int64,
	entityType string,
	entityID *int64,
	operation string,
	payload []byte,
) (*ProposeResult, error) {
	resp, err := s.ProposeOperation(tenantID, userID, entityType, entityID, operation, datatypes.JSON(payload))
	if err != nil {
		return nil, err
	}

	return &ProposeResult{
		ProposalID: resp.ID,
		Status:     resp.Status,
		Message:    resp.Message,
	}, nil
}

// GetPendingProposals returns pending proposals for a tenant
func (s *ProposalService) GetPendingProposals(tenantID int64, page, pageSize int) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, "created_at", "DESC")
	result, err := s.proposalRepo.FindPending(tenantID, params)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.ProposalResponse, len(result.Items))
	for i := range result.Items {
		items[i] = *proposalToResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// ApproveProposal approves and executes a proposal
func (s *ProposalService) ApproveProposal(proposalID, reviewerID int64) (*serializers.ProposalResponse, error) {
	proposal, err := s.proposalRepo.FindByID(proposalID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, ErrProposalNotFound
	}
	if err != nil {
		return nil, err
	}
	if proposal.Status != repositories.ProposalStatusPending {
		return nil, ErrProposalNotPending
	}
	if len(proposal.ValidationErrors) > 0 && string(proposal.ValidationErrors) != "null" {
		return nil, ErrProposalHasValidationErrors
	}

	// Mark approved
	if err := s.proposalRepo.UpdateStatus(proposalID, repositories.ProposalStatusApproved, &reviewerID); err != nil {
		return nil, err
	}

	// Execute the operation
	if err := s.executor.Execute(proposal.EntityType, proposal.Operation, proposal.EntityID, proposal.Payload, proposal.TenantID, reviewerID); err != nil {
		_ = s.proposalRepo.MarkFailed(proposalID, err.Error())
		return nil, ErrProposalExecutionFailed
	}

	if err := s.proposalRepo.MarkExecuted(proposalID); err != nil {
		return nil, err
	}

	proposal, _ = s.proposalRepo.FindByID(proposalID)
	return proposalToResponse(proposal), nil
}

// RejectProposal rejects a proposal
func (s *ProposalService) RejectProposal(proposalID, reviewerID int64) (*serializers.ProposalResponse, error) {
	proposal, err := s.proposalRepo.FindByID(proposalID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, ErrProposalNotFound
	}
	if err != nil {
		return nil, err
	}
	if proposal.Status != repositories.ProposalStatusPending {
		return nil, ErrProposalNotPending
	}

	if err := s.proposalRepo.UpdateStatus(proposalID, repositories.ProposalStatusRejected, &reviewerID); err != nil {
		return nil, err
	}

	proposal, _ = s.proposalRepo.FindByID(proposalID)
	return proposalToResponse(proposal), nil
}

// BulkApprove approves multiple proposals
func (s *ProposalService) BulkApprove(proposalIDs []int64, reviewerID int64) ([]serializers.ProposalResponse, []error) {
	results := make([]serializers.ProposalResponse, 0, len(proposalIDs))
	errs := make([]error, 0)

	for _, id := range proposalIDs {
		resp, err := s.ApproveProposal(id, reviewerID)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, *resp)
	}

	return results, errs
}

// BulkReject rejects multiple proposals
func (s *ProposalService) BulkReject(proposalIDs []int64, reviewerID int64) ([]serializers.ProposalResponse, []error) {
	results := make([]serializers.ProposalResponse, 0, len(proposalIDs))
	errs := make([]error, 0)

	for _, id := range proposalIDs {
		resp, err := s.RejectProposal(id, reviewerID)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, *resp)
	}

	return results, errs
}

// UpdateProposalPayload updates a proposal's payload and re-validates
func (s *ProposalService) UpdateProposalPayload(proposalID int64, payload []byte) (*serializers.ProposalResponse, error) {
	proposal, err := s.proposalRepo.FindByID(proposalID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, ErrProposalNotFound
	}
	if err != nil {
		return nil, err
	}
	if proposal.Status != repositories.ProposalStatusPending {
		return nil, ErrProposalNotPending
	}

	validationErrors := validation.ValidatePayload(proposal.EntityType, proposal.Operation, payload)
	var validationErrorsJSON datatypes.JSON
	if len(validationErrors) > 0 {
		validationErrorsJSON, _ = json.Marshal(validationErrors)
	}

	if err := s.proposalRepo.UpdatePayload(proposalID, datatypes.JSON(payload), validationErrorsJSON); err != nil {
		return nil, err
	}

	proposal, _ = s.proposalRepo.FindByID(proposalID)
	return proposalToResponse(proposal), nil
}

func isValidEntityType(entityType string) bool {
	for _, e := range repositories.SupportedEntities {
		if e == entityType {
			return true
		}
	}
	return false
}

func isValidOperation(operation string) bool {
	if operation == repositories.OperationCreate {
		return true
	}
	if operation == repositories.OperationUpdate {
		return true
	}
	if operation == repositories.OperationDelete {
		return true
	}
	return false
}

func proposalToResponse(p *repositories.ProposalDTO) *serializers.ProposalResponse {
	return &serializers.ProposalResponse{
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
