package tools

import (
	"context"
	"fmt"
	"log"
	"strings"

	"agent/internal/middleware"
	"agent/internal/services"
)

// PermissionChecker checks if the user in context has the specified permission
type PermissionChecker interface {
	CanAccess(ctx context.Context, permission string) bool
}

// ProposalCreator creates proposals for CUD operations
type ProposalCreator interface {
	ProposeOperationBytes(tenantID, userID int64, entityType string, entityID *int64, operation string, payload []byte) (*services.ProposeResult, error)
}

// EntityService provides generic entity operations
type EntityService interface {
	ListEntities(entityType string, limit int) ([]map[string]interface{}, error)
	SearchEntities(entityType, query string, limit int) ([]map[string]interface{}, error)
}

// CRMTools holds CRM-related tools for the agent
type CRMTools struct {
	entityService   EntityService
	permChecker     PermissionChecker
	proposalCreator ProposalCreator
}

// NewCRMTools creates CRM tools with service dependencies
func NewCRMTools(
	entityService EntityService,
	permChecker PermissionChecker,
	proposalCreator ProposalCreator,
) *CRMTools {
	return &CRMTools{
		entityService:   entityService,
		permChecker:     permChecker,
		proposalCreator: proposalCreator,
	}
}

// --- Tool Parameter Structs ---

// CRMLookupParams defines parameters for CRM entity lookup
type CRMLookupParams struct {
	EntityType string `json:"entity_type" jsonschema:"required,The type of CRM entity to search (e.g., individual, organization, job, deal, project)"`
	Query      string `json:"query,omitempty" jsonschema:"Search term for entity lookup (name, email, etc.)"`
}

// CRMListParams defines parameters for listing entities
type CRMListParams struct {
	EntityType string `json:"entity_type" jsonschema:"required,The type of CRM entity to list (e.g., individual, organization, job, deal, project)"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of results to return (default 20)"`
}

// CRMResult is the structured result of CRM operations
type CRMResult struct {
	Entities []map[string]interface{} `json:"entities"`
	Count    int                      `json:"count"`
	Error    string                   `json:"error,omitempty"`
}

// --- Tool Functions ---

// checkPermission returns an error result if user lacks the permission
func (t *CRMTools) checkPermission(ctx context.Context, permission string) *CRMResult {
	if t.permChecker == nil {
		return nil
	}
	if t.permChecker.CanAccess(ctx, permission) {
		return nil
	}
	log.Printf("🚫 Permission denied: %s", permission)
	return &CRMResult{Error: fmt.Sprintf("Permission denied: %s", permission)}
}

// normalizeEntityType converts entity type to permission resource
func normalizeEntityType(entityType string) string {
	resource := strings.ToLower(entityType)
	if resource == "lead" {
		return "job"
	}
	return resource
}

// LookupCtx searches for CRM entities by type and query
func (t *CRMTools) LookupCtx(ctx context.Context, params CRMLookupParams) (*CRMResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("🔍 crm_lookup entity_type='%s', query='%s'", entityType, params.Query)

	if denied := t.checkPermission(ctx, normalizeEntityType(entityType)+":read"); denied != nil {
		return denied, nil
	}

	if t.entityService == nil {
		return &CRMResult{Error: "Entity service not configured"}, nil
	}

	results, err := t.entityService.SearchEntities(entityType, params.Query, 20)
	if err != nil {
		log.Printf("❌ Lookup failed for %s: %v", entityType, err)
		return &CRMResult{Error: err.Error()}, nil
	}

	log.Printf("🔍 crm_lookup found %d %s", len(results), entityType)
	return &CRMResult{Entities: results, Count: len(results)}, nil
}

// ListCtx returns entities of a given type
func (t *CRMTools) ListCtx(ctx context.Context, params CRMListParams) (*CRMResult, error) {
	entityType := strings.ToLower(params.EntityType)
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	log.Printf("📋 crm_list entity_type='%s', limit=%d", entityType, limit)

	if denied := t.checkPermission(ctx, normalizeEntityType(entityType)+":read"); denied != nil {
		return denied, nil
	}

	if t.entityService == nil {
		return &CRMResult{Error: "Entity service not configured"}, nil
	}

	results, err := t.entityService.ListEntities(entityType, limit)
	if err != nil {
		log.Printf("❌ List failed for %s: %v", entityType, err)
		return &CRMResult{Error: err.Error()}, nil
	}

	log.Printf("📋 crm_list found %d %s", len(results), entityType)
	return &CRMResult{Entities: results, Count: len(results)}, nil
}

// --- Propose CUD Operations ---

// CRMProposeRecordCreateParams defines parameters for proposing record creation
type CRMProposeRecordCreateParams struct {
	EntityType string `json:"entity_type" jsonschema:"required,description=The type of CRM entity to create a record for"`
	DataJSON   string `json:"data_json" jsonschema:"required,description=The record data as a JSON string e.g. {\"title\":\"Engineer\"}"`
}

// CRMProposeRecordUpdateParams defines parameters for proposing record update
type CRMProposeRecordUpdateParams struct {
	EntityType string `json:"entity_type" jsonschema:"required,description=The type of CRM entity"`
	RecordID   int64  `json:"record_id" jsonschema:"required,description=The ID of the record to update"`
	DataJSON   string `json:"data_json" jsonschema:"required,description=The fields to update as a JSON string"`
}

// CRMProposeRecordDeleteParams defines parameters for proposing record deletion
type CRMProposeRecordDeleteParams struct {
	EntityType string `json:"entity_type" jsonschema:"required,The type of CRM entity"`
	RecordID   int64  `json:"record_id" jsonschema:"required,The ID of the record to delete"`
}

// CRMProposeResult is the result of a propose operation
type CRMProposeResult struct {
	Success    bool   `json:"success"`
	ProposalID int64  `json:"proposal_id,omitempty"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// ProposeRecordCreateCtx proposes creating a new CRM record (queued for approval)
func (t *CRMTools) ProposeRecordCreateCtx(ctx context.Context, params CRMProposeRecordCreateParams) (*CRMProposeResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_create entity_type='%s'", entityType)

	if denied := t.checkPermission(ctx, entityType+":create"); denied != nil {
		return &CRMProposeResult{Success: false, Message: denied.Error}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeResult{Success: false, Message: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, nil, "create", []byte(params.DataJSON))
	if err != nil {
		log.Printf("❌ ProposeCreate failed: %v", err)
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	log.Printf("✅ ProposeCreate succeeded: proposal_id=%d, status=%s", result.ProposalID, result.Status)
	return &CRMProposeResult{
		Success:    true,
		ProposalID: result.ProposalID,
		Status:     result.Status,
		Message:    result.Message,
	}, nil
}

// ProposeRecordUpdateCtx proposes updating a CRM record (queued for approval)
func (t *CRMTools) ProposeRecordUpdateCtx(ctx context.Context, params CRMProposeRecordUpdateParams) (*CRMProposeResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_record_update entity_type='%s', record_id=%d", entityType, params.RecordID)

	if denied := t.checkPermission(ctx, entityType+":update"); denied != nil {
		return &CRMProposeResult{Success: false, Message: denied.Error}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeResult{Success: false, Message: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	recordID := params.RecordID
	result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, &recordID, "update", []byte(params.DataJSON))
	if err != nil {
		log.Printf("❌ ProposeRecordUpdate failed: %v", err)
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	log.Printf("✅ ProposeRecordUpdate succeeded: proposal_id=%d, status=%s", result.ProposalID, result.Status)
	return &CRMProposeResult{
		Success:    true,
		ProposalID: result.ProposalID,
		Status:     result.Status,
		Message:    result.Message,
	}, nil
}

// ProposeRecordDeleteCtx proposes deleting a CRM record (queued for approval)
func (t *CRMTools) ProposeRecordDeleteCtx(ctx context.Context, params CRMProposeRecordDeleteParams) (*CRMProposeResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_record_delete entity_type='%s', record_id=%d", entityType, params.RecordID)

	if denied := t.checkPermission(ctx, entityType+":delete"); denied != nil {
		return &CRMProposeResult{Success: false, Message: denied.Error}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeResult{Success: false, Message: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	recordID := params.RecordID
	result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, &recordID, "delete", nil)
	if err != nil {
		log.Printf("❌ ProposeRecordDelete failed: %v", err)
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	log.Printf("✅ ProposeRecordDelete succeeded: proposal_id=%d, status=%s", result.ProposalID, result.Status)
	return &CRMProposeResult{
		Success:    true,
		ProposalID: result.ProposalID,
		Status:     result.Status,
		Message:    result.Message,
	}, nil
}

func (t *CRMTools) getContextIDs(ctx context.Context) (tenantID, userID int64, err error) {
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		return 0, 0, fmt.Errorf("tenant_id not found in context")
	}

	userID, ok = middleware.GetUserID(ctx)
	if !ok {
		return 0, 0, fmt.Errorf("user_id not found in context")
	}

	return tenantID, userID, nil
}
