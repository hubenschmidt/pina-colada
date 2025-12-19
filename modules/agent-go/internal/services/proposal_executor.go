package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pina-colada-co/agent-go/internal/schemas"
	"gorm.io/datatypes"
)

var ErrOperationNotSupported = errors.New("operation not supported for entity type")

// CRMProposalExecutor executes approved proposals on CRM entities
type CRMProposalExecutor struct {
	contactService *ContactService
	orgService     *OrganizationService
	indService     *IndividualService
	noteService    *NoteService
	taskService    *TaskService
	accountService *AccountService
}

// NewCRMProposalExecutor creates a new CRM proposal executor
func NewCRMProposalExecutor(
	contactService *ContactService,
	orgService *OrganizationService,
	indService *IndividualService,
	noteService *NoteService,
	taskService *TaskService,
	accountService *AccountService,
) *CRMProposalExecutor {
	return &CRMProposalExecutor{
		contactService: contactService,
		orgService:     orgService,
		indService:     indService,
		noteService:    noteService,
		taskService:    taskService,
		accountService: accountService,
	}
}

// Execute executes a proposal operation
func (e *CRMProposalExecutor) Execute(entityType, operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if entityType == "contact" {
		return e.executeContact(operation, entityID, payload, userID)
	}
	if entityType == "organization" {
		return e.executeOrganization(operation, entityID, payload, userID)
	}
	if entityType == "individual" {
		return e.executeIndividual(operation, entityID, payload, userID)
	}
	if entityType == "note" {
		return e.executeNote(operation, entityID, payload, tenantID, userID)
	}
	if entityType == "task" {
		return e.executeTask(operation, entityID, payload, tenantID, userID)
	}
	return fmt.Errorf("%w: %s", ErrOperationNotSupported, entityType)
}

func (e *CRMProposalExecutor) executeContact(operation string, entityID *int64, payload datatypes.JSON, userID int64) error {
	if operation == "create" {
		return e.createContact(payload, userID)
	}
	if operation == "update" {
		return e.updateContact(entityID, payload, userID)
	}
	if operation == "delete" {
		return e.deleteContact(entityID)
	}
	return ErrOperationNotSupported
}

func (e *CRMProposalExecutor) createContact(payload datatypes.JSON, userID int64) error {
	var input schemas.ContactCreate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse contact data: %w", err)
	}
	_, err := e.contactService.CreateContact(input, userID)
	return err
}

func (e *CRMProposalExecutor) updateContact(entityID *int64, payload datatypes.JSON, userID int64) error {
	if entityID == nil {
		return errors.New("entity_id required for update")
	}
	var input schemas.ContactUpdate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse contact data: %w", err)
	}
	_, err := e.contactService.UpdateContact(*entityID, input, userID)
	return err
}

func (e *CRMProposalExecutor) deleteContact(entityID *int64) error {
	if entityID == nil {
		return errors.New("entity_id required for delete")
	}
	return e.contactService.DeleteContact(*entityID)
}

func (e *CRMProposalExecutor) executeOrganization(operation string, entityID *int64, payload datatypes.JSON, userID int64) error {
	if operation == "create" {
		return e.createOrganization(payload, userID)
	}
	if operation == "update" {
		return e.updateOrganization(entityID, payload, userID)
	}
	if operation == "delete" {
		return e.deleteOrganization(entityID)
	}
	return ErrOperationNotSupported
}

func (e *CRMProposalExecutor) createOrganization(payload datatypes.JSON, userID int64) error {
	var input schemas.OrganizationCreate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse organization data: %w", err)
	}
	_, err := e.orgService.CreateOrganization(input, userID, nil)
	return err
}

func (e *CRMProposalExecutor) updateOrganization(entityID *int64, payload datatypes.JSON, userID int64) error {
	if entityID == nil {
		return errors.New("entity_id required for update")
	}
	var input schemas.OrganizationUpdate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse organization data: %w", err)
	}
	_, err := e.orgService.UpdateOrganization(*entityID, input, userID)
	return err
}

func (e *CRMProposalExecutor) deleteOrganization(entityID *int64) error {
	if entityID == nil {
		return errors.New("entity_id required for delete")
	}
	return e.orgService.DeleteOrganization(*entityID)
}

func (e *CRMProposalExecutor) executeIndividual(operation string, entityID *int64, payload datatypes.JSON, userID int64) error {
	if operation == "create" {
		return e.createIndividual(payload, userID)
	}
	if operation == "update" {
		return e.updateIndividual(entityID, payload, userID)
	}
	if operation == "delete" {
		return e.deleteIndividual(entityID)
	}
	return ErrOperationNotSupported
}

func (e *CRMProposalExecutor) createIndividual(payload datatypes.JSON, userID int64) error {
	var input schemas.IndividualCreate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse individual data: %w", err)
	}
	_, err := e.indService.CreateIndividual(input, nil, userID)
	return err
}

func (e *CRMProposalExecutor) updateIndividual(entityID *int64, payload datatypes.JSON, userID int64) error {
	if entityID == nil {
		return errors.New("entity_id required for update")
	}
	var input IndividualUpdateInput
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse individual data: %w", err)
	}
	_, err := e.indService.UpdateIndividual(*entityID, input, userID)
	return err
}

func (e *CRMProposalExecutor) deleteIndividual(entityID *int64) error {
	if entityID == nil {
		return errors.New("entity_id required for delete")
	}
	return e.indService.DeleteIndividual(*entityID)
}

func (e *CRMProposalExecutor) executeNote(operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if operation == "create" {
		return e.createNote(payload, tenantID, userID)
	}
	if operation == "update" {
		return e.updateNote(entityID, payload, userID)
	}
	if operation == "delete" {
		return e.deleteNote(entityID)
	}
	return ErrOperationNotSupported
}

func (e *CRMProposalExecutor) createNote(payload datatypes.JSON, tenantID, userID int64) error {
	var input schemas.NoteCreate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse note data: %w", err)
	}
	_, err := e.noteService.CreateNote(NoteCreateInput{
		TenantID:   tenantID,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Content:    input.Content,
		UserID:     userID,
	})
	return err
}

func (e *CRMProposalExecutor) updateNote(entityID *int64, payload datatypes.JSON, userID int64) error {
	if entityID == nil {
		return errors.New("entity_id required for update")
	}
	var input schemas.NoteUpdate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse note data: %w", err)
	}
	_, err := e.noteService.UpdateNote(*entityID, input.Content, userID)
	return err
}

func (e *CRMProposalExecutor) deleteNote(entityID *int64) error {
	if entityID == nil {
		return errors.New("entity_id required for delete")
	}
	return e.noteService.DeleteNote(*entityID)
}

func (e *CRMProposalExecutor) executeTask(operation string, entityID *int64, payload datatypes.JSON, tenantID, userID int64) error {
	if operation == "create" {
		return e.createTask(payload, tenantID, userID)
	}
	if operation == "update" {
		return e.updateTask(entityID, payload, userID)
	}
	if operation == "delete" {
		return e.deleteTask(entityID)
	}
	return ErrOperationNotSupported
}

func (e *CRMProposalExecutor) createTask(payload datatypes.JSON, tenantID, userID int64) error {
	var input schemas.TaskCreate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse task data: %w", err)
	}
	_, err := e.taskService.CreateTask(input, &tenantID, userID)
	return err
}

func (e *CRMProposalExecutor) updateTask(entityID *int64, payload datatypes.JSON, userID int64) error {
	if entityID == nil {
		return errors.New("entity_id required for update")
	}
	var input schemas.TaskUpdate
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to parse task data: %w", err)
	}
	_, err := e.taskService.UpdateTask(*entityID, input, userID)
	return err
}

func (e *CRMProposalExecutor) deleteTask(entityID *int64) error {
	if entityID == nil {
		return errors.New("entity_id required for delete")
	}
	return e.taskService.DeleteTask(*entityID)
}

// Ensure CRMProposalExecutor implements ProposalExecutor
var _ ProposalExecutor = (*CRMProposalExecutor)(nil)
