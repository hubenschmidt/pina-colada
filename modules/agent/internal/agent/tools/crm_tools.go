package tools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

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

// maxConcurrentWorkers limits concurrent DB operations for resource-constrained environments
const maxConcurrentWorkers = 20

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

// CRMProposeBatchCreateParams defines parameters for batch proposing record creation
type CRMProposeBatchCreateParams struct {
	EntityType string   `json:"entity_type" jsonschema:"required,description=The type of CRM entity to create records for (e.g. job)"`
	ItemsJSON  []string `json:"items_json" jsonschema:"required,description=Array of record data as JSON strings. Each item is a JSON object string e.g. [{\"title\":\"Engineer\"},{\"title\":\"Manager\"}]"`
}

// CRMProposeBatchResult is the result of a batch propose operation
type CRMProposeBatchResult struct {
	Success      bool    `json:"success"`
	Created      int     `json:"created"`
	Failed       int     `json:"failed"`
	ProposalIDs  []int64 `json:"proposal_ids,omitempty"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// CRMProposeBatchUpdateItem represents a single update in a batch
type CRMProposeBatchUpdateItem struct {
	RecordID int64  `json:"record_id" jsonschema:"required,description=The ID of the record to update"`
	DataJSON string `json:"data_json" jsonschema:"required,description=The fields to update as a JSON string"`
}

// CRMProposeBatchUpdateParams defines parameters for batch proposing record updates
type CRMProposeBatchUpdateParams struct {
	EntityType string                      `json:"entity_type" jsonschema:"required,description=The type of CRM entity to update (e.g. job)"`
	Items      []CRMProposeBatchUpdateItem `json:"items" jsonschema:"required,description=Array of update items, each with record_id and data_json"`
}

// CRMProposeBulkUpdateAllParams defines parameters for bulk updating all records of a type
type CRMProposeBulkUpdateAllParams struct {
	EntityType string `json:"entity_type" jsonschema:"required,description=The type of CRM entity to update (e.g. job)"`
	DataJSON   string `json:"data_json" jsonschema:"required,description=The fields to update as a JSON string (applied to ALL records)"`
}

// ProposeRecordBatchCreateCtx proposes creating multiple CRM records in one call (concurrent)
func (t *CRMTools) ProposeRecordBatchCreateCtx(ctx context.Context, params CRMProposeBatchCreateParams) (*CRMProposeBatchResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_batch_create entity_type='%s', count=%d", entityType, len(params.ItemsJSON))

	if denied := t.checkPermission(ctx, entityType+":create"); denied != nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: denied.Error}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: err.Error()}, nil
	}

	type batchResult struct {
		index      int
		proposalID int64
		err        error
	}

	results := make(chan batchResult, len(params.ItemsJSON))
	sem := make(chan struct{}, maxConcurrentWorkers)
	var wg sync.WaitGroup

	for i, itemJSON := range params.ItemsJSON {
		wg.Add(1)
		go func(idx int, data string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, nil, "create", []byte(data))
			if err != nil {
				results <- batchResult{index: idx, err: err}
				return
			}
			results <- batchResult{index: idx, proposalID: result.ProposalID}
		}(i, itemJSON)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var proposalIDs []int64
	var created, failed int
	for r := range results {
		if r.err != nil {
			log.Printf("❌ Batch item %d failed: %v", r.index, r.err)
			failed++
			continue
		}
		proposalIDs = append(proposalIDs, r.proposalID)
		created++
	}

	log.Printf("✅ ProposeBatchCreate completed: created=%d, failed=%d", created, failed)
	return &CRMProposeBatchResult{
		Success:     failed == 0,
		Created:     created,
		Failed:      failed,
		ProposalIDs: proposalIDs,
	}, nil
}

// ProposeRecordBatchUpdateCtx proposes updating multiple CRM records in one call (concurrent)
func (t *CRMTools) ProposeRecordBatchUpdateCtx(ctx context.Context, params CRMProposeBatchUpdateParams) (*CRMProposeBatchResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_batch_update entity_type='%s', count=%d", entityType, len(params.Items))

	if denied := t.checkPermission(ctx, entityType+":update"); denied != nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: denied.Error}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: err.Error()}, nil
	}

	type batchResult struct {
		index      int
		proposalID int64
		err        error
	}

	results := make(chan batchResult, len(params.Items))
	sem := make(chan struct{}, maxConcurrentWorkers)
	var wg sync.WaitGroup

	for i, item := range params.Items {
		wg.Add(1)
		go func(idx int, recordID int64, data string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, &recordID, "update", []byte(data))
			if err != nil {
				results <- batchResult{index: idx, err: err}
				return
			}
			results <- batchResult{index: idx, proposalID: result.ProposalID}
		}(i, item.RecordID, item.DataJSON)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var proposalIDs []int64
	var created, failed int
	for r := range results {
		if r.err != nil {
			log.Printf("❌ Batch update item %d failed: %v", r.index, r.err)
			failed++
			continue
		}
		proposalIDs = append(proposalIDs, r.proposalID)
		created++
	}

	log.Printf("✅ ProposeBatchUpdate completed: updated=%d, failed=%d", created, failed)
	return &CRMProposeBatchResult{
		Success:     failed == 0,
		Created:     created,
		Failed:      failed,
		ProposalIDs: proposalIDs,
	}, nil
}

// ProposeBulkUpdateAllCtx proposes updating ALL records of a type with the same data
func (t *CRMTools) ProposeBulkUpdateAllCtx(ctx context.Context, params CRMProposeBulkUpdateAllParams) (*CRMProposeBatchResult, error) {
	entityType := strings.ToLower(params.EntityType)
	limit := 10000 // Internal limit, not exposed to agent
	log.Printf("📝 crm_propose_bulk_update_all entity_type='%s', limit=%d", entityType, limit)

	if denied := t.checkPermission(ctx, entityType+":update"); denied != nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: denied.Error}, nil
	}

	if t.proposalCreator == nil || t.entityService == nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: "Service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: err.Error()}, nil
	}

	// Fetch all entity IDs
	entities, err := t.entityService.ListEntities(entityType, limit)
	if err != nil {
		return &CRMProposeBatchResult{Success: false, ErrorMessage: err.Error()}, nil
	}

	log.Printf("📝 Found %d %s records to update", len(entities), entityType)

	type batchResult struct {
		index      int
		proposalID int64
		err        error
	}

	results := make(chan batchResult, len(entities))
	sem := make(chan struct{}, maxConcurrentWorkers) // Semaphore to limit concurrency
	var wg sync.WaitGroup

	for i, entity := range entities {
		idVal, ok := entity["id"]
		if !ok {
			continue
		}
		var recordID int64
		switch v := idVal.(type) {
		case float64:
			recordID = int64(v)
		case int64:
			recordID = v
		case int:
			recordID = int64(v)
		default:
			continue
		}

		wg.Add(1)
		go func(idx int, id int64) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, &id, "update", []byte(params.DataJSON))
			if err != nil {
				results <- batchResult{index: idx, err: err}
				return
			}
			results <- batchResult{index: idx, proposalID: result.ProposalID}
		}(i, recordID)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var proposalIDs []int64
	var created, failed int
	for r := range results {
		if r.err != nil {
			log.Printf("❌ Bulk update item %d failed: %v", r.index, r.err)
			failed++
			continue
		}
		proposalIDs = append(proposalIDs, r.proposalID)
		created++
	}

	log.Printf("✅ ProposeBulkUpdateAll completed: updated=%d, failed=%d", created, failed)
	return &CRMProposeBatchResult{
		Success:     failed == 0,
		Created:     created,
		Failed:      failed,
		ProposalIDs: proposalIDs,
	}, nil
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
