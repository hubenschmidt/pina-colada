package services

import (
	"encoding/json"
	"errors"
	"fmt"

	apperrors "agent/internal/errors"
	"agent/internal/repositories"
	"agent/internal/schemas"
	"agent/internal/serializers"
	"agent/internal/validation"
	"gorm.io/datatypes"
)

var (
	ErrProposalNotFound            = errors.New("proposal not found")
	ErrProposalNotPending          = errors.New("proposal is not in pending status")
	ErrProposalInvalidEntity       = errors.New("invalid entity type")
	ErrProposalInvalidOperation    = errors.New("invalid operation")
	ErrProposalExecutionFailed     = errors.New("proposal execution failed")
	ErrProposalHasValidationErrors = errors.New("proposal has validation errors - fix payload before approving")
	ErrOperationNotSupported       = errors.New("operation not supported for entity type")
)

// ProposeResult is a simplified result for tool callers
type ProposeResult struct {
	ProposalID int64
	Status     string
	Message    string
}

// ProposalService handles proposal business logic
type ProposalService struct {
	proposalRepo   *repositories.ProposalRepository
	approvalRepo   *repositories.ApprovalConfigRepository
	contactService *ContactService
	orgService     *OrganizationService
	indService     *IndividualService
	noteService    *NoteService
	taskService    *TaskService
	jobService     *JobService
	leadService    *LeadService
}

// NewProposalService creates a new proposal service
func NewProposalService(
	proposalRepo *repositories.ProposalRepository,
	approvalRepo *repositories.ApprovalConfigRepository,
	contactService *ContactService,
	orgService *OrganizationService,
	indService *IndividualService,
	noteService *NoteService,
	taskService *TaskService,
	jobService *JobService,
	leadService *LeadService,
) *ProposalService {
	return &ProposalService{
		proposalRepo:   proposalRepo,
		approvalRepo:   approvalRepo,
		contactService: contactService,
		orgService:     orgService,
		indService:     indService,
		noteService:    noteService,
		taskService:    taskService,
		jobService:     jobService,
		leadService:    leadService,
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

	if !requiresApproval {
		err := s.execute(entityType, operation, entityID, payload, tenantID, userID)
		if err != nil {
			return nil, err
		}
		return &serializers.ProposalResponse{
			Status:  repositories.ProposalStatusExecuted,
			Message: "operation executed immediately (approval not required)",
		}, nil
	}

	validationErrors := validation.ValidatePayload(entityType, operation, payload)
	var validationErrorsJSON datatypes.JSON
	if len(validationErrors) > 0 {
		validationErrorsJSON, _ = json.Marshal(validationErrors)
	}

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

	err = s.execute(proposal.EntityType, proposal.Operation, proposal.EntityID, proposal.Payload, proposal.TenantID, reviewerID)
	if err != nil {
		_ = s.proposalRepo.SetError(proposalID, err.Error())
		return nil, fmt.Errorf("%w: %s", ErrProposalExecutionFailed, err.Error())
	}

	err = s.proposalRepo.UpdateStatus(proposalID, repositories.ProposalStatusApproved, &reviewerID)
	if err != nil {
		return nil, err
	}

	err = s.proposalRepo.MarkExecuted(proposalID)
	if err != nil {
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

	err = s.proposalRepo.UpdateStatus(proposalID, repositories.ProposalStatusRejected, &reviewerID)
	if err != nil {
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
		resp, err := s.approveOne(id, reviewerID)
		results, errs = collectResult(results, errs, resp, err)
	}

	return results, errs
}

// BulkReject rejects multiple proposals
func (s *ProposalService) BulkReject(proposalIDs []int64, reviewerID int64) ([]serializers.ProposalResponse, []error) {
	results := make([]serializers.ProposalResponse, 0, len(proposalIDs))
	errs := make([]error, 0)

	for _, id := range proposalIDs {
		resp, err := s.rejectOne(id, reviewerID)
		results, errs = collectResult(results, errs, resp, err)
	}

	return results, errs
}

func (s *ProposalService) approveOne(id, reviewerID int64) (*serializers.ProposalResponse, error) {
	return s.ApproveProposal(id, reviewerID)
}

func (s *ProposalService) rejectOne(id, reviewerID int64) (*serializers.ProposalResponse, error) {
	return s.RejectProposal(id, reviewerID)
}

func collectResult(results []serializers.ProposalResponse, errs []error, resp *serializers.ProposalResponse, err error) ([]serializers.ProposalResponse, []error) {
	if err != nil {
		return results, append(errs, err)
	}
	return append(results, *resp), errs
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

	err = s.proposalRepo.UpdatePayload(proposalID, datatypes.JSON(payload), validationErrorsJSON)
	if err != nil {
		return nil, err
	}

	proposal, _ = s.proposalRepo.FindByID(proposalID)
	return proposalToResponse(proposal), nil
}

// execute calls the appropriate service based on entity type
func (s *ProposalService) execute(entityType, operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if entityType == "contact" {
		return s.executeContact(operation, entityID, payload, userID)
	}
	if entityType == "organization" {
		return s.executeOrganization(operation, entityID, payload, userID)
	}
	if entityType == "individual" {
		return s.executeIndividual(operation, entityID, payload, userID)
	}
	if entityType == "note" {
		return s.executeNote(operation, entityID, payload, tenantID, userID)
	}
	if entityType == "task" {
		return s.executeTask(operation, entityID, payload, tenantID, userID)
	}
	if entityType == "job" || entityType == "lead" {
		return s.executeJob(operation, entityID, payload, tenantID, userID)
	}
	if entityType == "opportunity" {
		return s.executeOpportunity(operation, entityID, payload, tenantID, userID)
	}
	if entityType == "partnership" {
		return s.executePartnership(operation, entityID, payload, tenantID, userID)
	}
	return fmt.Errorf("%w: %s", ErrOperationNotSupported, entityType)
}

func (s *ProposalService) executeContact(operation string, entityID *int64, payload datatypes.JSON, userID int64) error {
	if operation == "create" {
		return s.createContact(payload, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updateContact(*entityID, payload, userID)
	}
	if operation == "delete" {
		return s.contactService.DeleteContact(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createContact(payload datatypes.JSON, userID int64) error {
	var input schemas.ContactCreate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("contact", err)
	}
	_, err = s.contactService.CreateContact(input, userID)
	return err
}

func (s *ProposalService) updateContact(id int64, payload datatypes.JSON, userID int64) error {
	var input schemas.ContactUpdate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("contact", err)
	}
	_, err = s.contactService.UpdateContact(id, input, userID)
	return err
}

func (s *ProposalService) executeOrganization(operation string, entityID *int64, payload datatypes.JSON, userID int64) error {
	if operation == "create" {
		return s.createOrganization(payload, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updateOrganization(*entityID, payload, userID)
	}
	if operation == "delete" {
		return s.orgService.DeleteOrganization(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createOrganization(payload datatypes.JSON, userID int64) error {
	var input schemas.OrganizationCreate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("organization", err)
	}
	_, err = s.orgService.CreateOrganization(input, userID, nil)
	return err
}

func (s *ProposalService) updateOrganization(id int64, payload datatypes.JSON, userID int64) error {
	var input schemas.OrganizationUpdate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("organization", err)
	}
	_, err = s.orgService.UpdateOrganization(id, input, userID)
	return err
}

func (s *ProposalService) executeIndividual(operation string, entityID *int64, payload datatypes.JSON, userID int64) error {
	if operation == "create" {
		return s.createIndividual(payload, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updateIndividual(*entityID, payload, userID)
	}
	if operation == "delete" {
		return s.indService.DeleteIndividual(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createIndividual(payload datatypes.JSON, userID int64) error {
	var input schemas.IndividualCreate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("individual", err)
	}
	_, err = s.indService.CreateIndividual(input, nil, userID)
	return err
}

func (s *ProposalService) updateIndividual(id int64, payload datatypes.JSON, userID int64) error {
	var input IndividualUpdateInput
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("individual", err)
	}
	_, err = s.indService.UpdateIndividual(id, input, userID)
	return err
}

func (s *ProposalService) executeNote(operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if operation == "create" {
		return s.createNote(payload, tenantID, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updateNote(*entityID, payload, userID)
	}
	if operation == "delete" {
		return s.noteService.DeleteNote(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createNote(payload datatypes.JSON, tenantID, userID int64) error {
	var input schemas.NoteCreate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("note", err)
	}
	_, err = s.noteService.CreateNote(NoteCreateInput{
		TenantID:   tenantID,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Content:    input.Content,
		UserID:     userID,
	})
	return err
}

func (s *ProposalService) updateNote(id int64, payload datatypes.JSON, userID int64) error {
	var input schemas.NoteUpdate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("note", err)
	}
	_, err = s.noteService.UpdateNote(id, input.Content, userID)
	return err
}

func (s *ProposalService) executeTask(operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if operation == "create" {
		return s.createTask(payload, tenantID, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updateTask(*entityID, payload, userID)
	}
	if operation == "delete" {
		return s.taskService.DeleteTask(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createTask(payload datatypes.JSON, tenantID, userID int64) error {
	var input schemas.TaskCreate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("task", err)
	}
	_, err = s.taskService.CreateTask(input, &tenantID, userID)
	return err
}

func (s *ProposalService) updateTask(id int64, payload datatypes.JSON, userID int64) error {
	var input schemas.TaskUpdate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("task", err)
	}
	_, err = s.taskService.UpdateTask(id, input, userID)
	return err
}

func (s *ProposalService) executeJob(operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if operation == "create" {
		return s.createJob(payload, tenantID, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updateJob(*entityID, payload, tenantID, userID)
	}
	if operation == "delete" {
		return s.jobService.DeleteJob(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createJob(payload datatypes.JSON, tenantID, userID int64) error {
	// Normalize payload: map common LLM field names to schema field names
	normalized := normalizeJobPayload(payload)

	var input schemas.JobCreate
	err := json.Unmarshal(normalized, &input)
	if err != nil {
		return parseError("job", err)
	}
	_, err = s.jobService.CreateJob(input, &tenantID, userID)
	return err
}

// normalizeJobPayload maps LLM-generated field names to schema field names
func normalizeJobPayload(payload datatypes.JSON) datatypes.JSON {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return payload
	}

	// Map "title" -> "job_title" if job_title not present
	if _, hasJobTitle := data["job_title"]; !hasJobTitle {
		if title, hasTitle := data["title"]; hasTitle {
			data["job_title"] = title
			delete(data, "title")
		}
	}

	// Map "company" -> "account" if account not present
	if _, hasAccount := data["account"]; !hasAccount {
		if company, hasCompany := data["company"]; hasCompany {
			data["account"] = company
			delete(data, "company")
		}
	}

	// Map "url" -> "job_url" if job_url not present
	if _, hasJobURL := data["job_url"]; !hasJobURL {
		if url, hasURL := data["url"]; hasURL {
			data["job_url"] = url
			delete(data, "url")
		}
	}

	// Default status to "Lead" if not present
	if _, hasStatus := data["status"]; !hasStatus {
		data["status"] = "Lead"
	}

	normalized, _ := json.Marshal(data)
	return datatypes.JSON(normalized)
}

func (s *ProposalService) updateJob(id int64, payload datatypes.JSON, tenantID, userID int64) error {
	var input schemas.JobUpdate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("job", err)
	}
	_, err = s.jobService.UpdateJob(id, input, &tenantID, userID)
	return err
}

func (s *ProposalService) executeOpportunity(operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if operation == "create" {
		return s.createOpportunity(payload, tenantID, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updateOpportunity(*entityID, payload, userID)
	}
	if operation == "delete" {
		return s.leadService.DeleteOpportunity(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createOpportunity(payload datatypes.JSON, tenantID, userID int64) error {
	var input schemas.OpportunityCreate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("opportunity", err)
	}
	_, err = s.leadService.CreateOpportunity(input, &tenantID, userID)
	return err
}

func (s *ProposalService) updateOpportunity(id int64, payload datatypes.JSON, userID int64) error {
	var input schemas.OpportunityUpdate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("opportunity", err)
	}
	_, err = s.leadService.UpdateOpportunity(id, input, userID)
	return err
}

func (s *ProposalService) executePartnership(operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if operation == "create" {
		return s.createPartnership(payload, tenantID, userID)
	}
	if entityID == nil {
		return errEntityIDRequired(operation)
	}
	if operation == "update" {
		return s.updatePartnership(*entityID, payload, userID)
	}
	if operation == "delete" {
		return s.leadService.DeletePartnership(*entityID)
	}
	return ErrOperationNotSupported
}

func (s *ProposalService) createPartnership(payload datatypes.JSON, tenantID, userID int64) error {
	var input schemas.PartnershipCreate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("partnership", err)
	}
	_, err = s.leadService.CreatePartnership(input, &tenantID, userID)
	return err
}

func (s *ProposalService) updatePartnership(id int64, payload datatypes.JSON, userID int64) error {
	var input schemas.PartnershipUpdate
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return parseError("partnership", err)
	}
	_, err = s.leadService.UpdatePartnership(id, input, userID)
	return err
}

func parseError(entity string, err error) error {
	return fmt.Errorf("failed to parse %s data: %w", entity, err)
}

func errEntityIDRequired(operation string) error {
	return fmt.Errorf("entity_id required for %s", operation)
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
