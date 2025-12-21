package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// PermissionChecker checks if the user in context has the specified permission
type PermissionChecker interface {
	CanAccess(ctx context.Context, permission string) bool
}

// ProposalCreator creates proposals for CUD operations
type ProposalCreator interface {
	ProposeOperationBytes(tenantID, userID int64, entityType string, entityID *int64, operation string, payload []byte) (*services.ProposeResult, error)
}

// GenericEntityLister lists entities from any supported table
type GenericEntityLister interface {
	ListEntities(entityType string, limit int) ([]map[string]interface{}, error)
}

// CRMTools holds CRM-related tools for the agent
type CRMTools struct {
	indService      *services.IndividualService
	orgService      *services.OrganizationService
	contactService  *services.ContactService
	jobService      *services.JobService
	accountService  *services.AccountService
	entityLister    GenericEntityLister
	permChecker     PermissionChecker
	proposalCreator ProposalCreator
}

// NewCRMTools creates CRM tools with service dependencies
func NewCRMTools(
	indService *services.IndividualService,
	orgService *services.OrganizationService,
	contactService *services.ContactService,
	jobService *services.JobService,
	accountService *services.AccountService,
	entityLister GenericEntityLister,
	permChecker PermissionChecker,
	proposalCreator ProposalCreator,
) *CRMTools {
	return &CRMTools{
		indService:      indService,
		orgService:      orgService,
		contactService:  contactService,
		jobService:      jobService,
		accountService:  accountService,
		entityLister:    entityLister,
		permChecker:     permChecker,
		proposalCreator: proposalCreator,
	}
}

// --- Tool Parameter Structs ---

// CRMLookupParams defines parameters for CRM entity lookup
type CRMLookupParams struct {
	EntityType string   `json:"entity_type" jsonschema:"The type of CRM entity to search: individual, organization, contact, account, job, or lead"`
	Query      string   `json:"query,omitempty" jsonschema:"Search term for entity lookup (name, email, etc.)"`
	Status     []string `json:"status,omitempty" jsonschema:"Filter by status names (for job/lead only, e.g., applied, interviewing)"`
}

// CRMLookupResult is the result of a CRM lookup
type CRMLookupResult struct {
	Results string `json:"results"`
	Count   int    `json:"count"`
}

// CRMCountParams defines parameters for counting entities
type CRMCountParams struct {
	EntityType string `json:"entity_type" jsonschema:"The type of CRM entity to count: individual, organization, account, or contact"`
}

// CRMCountResult is the result of a count operation
type CRMCountResult struct {
	EntityType string `json:"entity_type"`
	Count      int    `json:"count"`
}

// CRMListParams defines parameters for listing entities
type CRMListParams struct {
	EntityType string `json:"entity_type" jsonschema:"The type of CRM entity to list: individual, organization, contact, account, job, or lead"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of results to return (default 20)"`
}

// CRMListResult is the result of a list operation
type CRMListResult struct {
	Results string `json:"results"`
	Count   int    `json:"count"`
}

// CRMStatusesParams defines parameters for listing available statuses
type CRMStatusesParams struct{}

// CRMStatusesResult is the result of listing statuses
type CRMStatusesResult struct {
	Results string `json:"results"`
	Count   int    `json:"count"`
}

// --- Tool Functions ---

// checkPermission returns a permission denied result if user lacks the permission
func (t *CRMTools) checkPermission(ctx context.Context, permission string) *CRMLookupResult {
	if t.permChecker == nil {
		return nil // No checker configured, allow all
	}
	if t.permChecker.CanAccess(ctx, permission) {
		return nil
	}
	log.Printf("🚫 Permission denied: %s", permission)
	return &CRMLookupResult{Results: fmt.Sprintf("Permission denied: %s", permission)}
}

// LookupCtx searches for CRM entities by type and query.
func (t *CRMTools) LookupCtx(ctx context.Context, params CRMLookupParams) (*CRMLookupResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("🔍 crm_lookup called with entity_type='%s', query='%s', status=%v", entityType, params.Query, params.Status)

	// Map entity type to permission resource
	resource := entityType
	if resource == "lead" {
		resource = "job"
	}

	if denied := t.checkPermission(ctx, resource+":read"); denied != nil {
		return denied, nil
	}

	if entityType == "individual" {
		return t.lookupIndividual(params.Query)
	}
	if entityType == "organization" {
		return t.lookupOrganization(params.Query)
	}
	if entityType == "contact" {
		return t.lookupContact(params.Query)
	}
	if entityType == "account" {
		return t.lookupAccount(params.Query)
	}
	if entityType == "job" || entityType == "lead" {
		return t.lookupJobLeads(params.Status)
	}
	log.Printf("⚠️  Unknown entity type: %s", entityType)
	return &CRMLookupResult{
		Results: fmt.Sprintf("Unknown entity type: %s. Supported: individual, organization, contact, account, job, lead", entityType),
	}, nil
}

func (t *CRMTools) lookupIndividual(query string) (*CRMLookupResult, error) {
	if t.indService == nil {
		log.Printf("⚠️ IndividualService not configured")
		return &CRMLookupResult{Results: "CRM service not configured. Unable to look up individuals."}, nil
	}
	results, err := t.indService.SearchIndividuals(query, nil, 10)
	if err != nil {
		log.Printf("❌ Individual lookup failed: %v", err)
		return &CRMLookupResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("🔍 crm_lookup found %d individuals", len(results))
	if len(results) == 0 {
		return &CRMLookupResult{Results: fmt.Sprintf("No individuals found matching '%s'", query)}, nil
	}
	var lines []string
	for _, ind := range results {
		lines = append(lines, formatIndividual(ind))
	}
	return &CRMLookupResult{
		Results: fmt.Sprintf("Found %d individuals:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) lookupOrganization(query string) (*CRMLookupResult, error) {
	if t.orgService == nil {
		log.Printf("⚠️ OrganizationService not configured")
		return &CRMLookupResult{Results: "CRM service not configured. Unable to look up organizations."}, nil
	}
	results, err := t.orgService.SearchOrganizations(query, nil, 10)
	if err != nil {
		log.Printf("❌ Organization lookup failed: %v", err)
		return &CRMLookupResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("🔍 crm_lookup found %d organizations", len(results))
	if len(results) == 0 {
		return &CRMLookupResult{Results: fmt.Sprintf("No organizations found matching '%s'", query)}, nil
	}
	var lines []string
	for _, org := range results {
		lines = append(lines, fmt.Sprintf("- %s (id=%d)", org.Name, org.ID))
	}
	return &CRMLookupResult{
		Results: fmt.Sprintf("Found %d organizations:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) lookupContact(query string) (*CRMLookupResult, error) {
	if t.contactService == nil {
		log.Printf("⚠️ ContactService not configured")
		return &CRMLookupResult{Results: "CRM service not configured. Unable to look up contacts."}, nil
	}
	results, err := t.contactService.SearchContacts(query, 10)
	if err != nil {
		log.Printf("❌ Contact lookup failed: %v", err)
		return &CRMLookupResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("🔍 crm_lookup found %d contacts", len(results))
	if len(results) == 0 {
		return &CRMLookupResult{Results: fmt.Sprintf("No contacts found matching '%s'", query)}, nil
	}
	var lines []string
	for _, c := range results {
		lines = append(lines, formatContact(c))
	}
	return &CRMLookupResult{
		Results: fmt.Sprintf("Found %d contacts:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) lookupAccount(query string) (*CRMLookupResult, error) {
	if t.accountService == nil {
		log.Printf("⚠️ AccountService not configured")
		return &CRMLookupResult{Results: "CRM service not configured. Unable to look up accounts."}, nil
	}
	results, err := t.accountService.SearchAccounts(query, nil, 10)
	if err != nil {
		log.Printf("❌ Account lookup failed: %v", err)
		return &CRMLookupResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("🔍 crm_lookup found %d accounts", len(results))
	if len(results) == 0 {
		return &CRMLookupResult{Results: fmt.Sprintf("No accounts found matching '%s'", query)}, nil
	}
	var lines []string
	for _, a := range results {
		lines = append(lines, fmt.Sprintf("- %s (type=%s, id=%d)", a.Name, a.Type, a.ID))
	}
	return &CRMLookupResult{
		Results: fmt.Sprintf("Found %d accounts:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) lookupJobLeads(statusNames []string) (*CRMLookupResult, error) {
	if t.jobService == nil {
		log.Printf("⚠️ JobService not configured")
		return &CRMLookupResult{Results: "CRM service not configured. Unable to look up job leads."}, nil
	}
	results, err := t.jobService.GetLeads(statusNames, nil)
	if err != nil {
		log.Printf("❌ Job leads lookup failed: %v", err)
		return &CRMLookupResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("🔍 crm_lookup found %d job leads", len(results))
	statusMsg := "all statuses"
	if len(statusNames) > 0 {
		statusMsg = fmt.Sprintf("status %v", statusNames)
	}
	if len(results) == 0 {
		return &CRMLookupResult{Results: fmt.Sprintf("No job leads found with %s.", statusMsg)}, nil
	}
	var lines []string
	for _, j := range results {
		lines = append(lines, formatJobLead(j))
	}
	return &CRMLookupResult{
		Results: fmt.Sprintf("Found %d job leads with %s:\n%s", len(results), statusMsg, strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

// ListStatusesCtx returns available job lead statuses
func (t *CRMTools) ListStatusesCtx(ctx context.Context, params CRMStatusesParams) (*CRMStatusesResult, error) {
	if denied := t.checkPermission(ctx, "job:read"); denied != nil {
		return &CRMStatusesResult{Results: denied.Results}, nil
	}
	if t.jobService == nil {
		log.Printf("⚠️ JobService not configured")
		return &CRMStatusesResult{Results: "Job service not configured."}, nil
	}
	log.Printf("📋 crm_statuses called")
	statuses, err := t.jobService.GetLeadStatuses()
	if err != nil {
		log.Printf("❌ Lead statuses lookup failed: %v", err)
		return &CRMStatusesResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	if len(statuses) == 0 {
		return &CRMStatusesResult{Results: "No job lead statuses configured."}, nil
	}
	log.Printf("📋 crm_statuses found %d statuses", len(statuses))
	var lines []string
	for _, s := range statuses {
		terminal := ""
		if s.IsTerminal {
			terminal = " [terminal]"
		}
		lines = append(lines, fmt.Sprintf("- %s%s", s.Name, terminal))
	}
	return &CRMStatusesResult{
		Results: fmt.Sprintf("Available job lead statuses:\n%s", strings.Join(lines, "\n")),
		Count:   len(statuses),
	}, nil
}

// ListCtx returns entities of a given type.
func (t *CRMTools) ListCtx(ctx context.Context, params CRMListParams) (*CRMListResult, error) {
	entityType := strings.ToLower(params.EntityType)
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	log.Printf("📋 crm_list called with entity_type='%s', limit=%d", entityType, limit)

	// Map entity type to permission resource (lead uses job permission)
	resource := entityType
	if resource == "lead" {
		resource = "job"
	}

	if denied := t.checkPermission(ctx, resource+":read"); denied != nil {
		return &CRMListResult{Results: denied.Results}, nil
	}

	if entityType == "individual" {
		return t.listIndividuals(limit)
	}
	if entityType == "organization" {
		return t.listOrganizations(limit)
	}
	if entityType == "contact" {
		return t.listContacts(limit)
	}
	if entityType == "account" {
		return t.listAccounts(limit)
	}
	if entityType == "job" || entityType == "lead" {
		return t.listJobs(limit)
	}
	// Fall through to generic database query for other entity types
	return t.listGeneric(entityType, limit)
}

func (t *CRMTools) listIndividuals(limit int) (*CRMListResult, error) {
	if t.indService == nil {
		log.Printf("⚠️ IndividualService not configured")
		return &CRMListResult{Results: "CRM service not configured. Unable to list individuals."}, nil
	}
	results, err := t.indService.SearchIndividuals("", nil, limit)
	if err != nil {
		log.Printf("❌ Individual list failed: %v", err)
		return &CRMListResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("📋 crm_list found %d individuals", len(results))
	if len(results) == 0 {
		return &CRMListResult{Results: "No individuals found."}, nil
	}
	var lines []string
	for _, ind := range results {
		lines = append(lines, formatIndividual(ind))
	}
	return &CRMListResult{
		Results: fmt.Sprintf("Found %d individuals:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) listOrganizations(limit int) (*CRMListResult, error) {
	if t.orgService == nil {
		log.Printf("⚠️ OrganizationService not configured")
		return &CRMListResult{Results: "CRM service not configured. Unable to list organizations."}, nil
	}
	results, err := t.orgService.SearchOrganizations("", nil, limit)
	if err != nil {
		log.Printf("❌ Organization list failed: %v", err)
		return &CRMListResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("📋 crm_list found %d organizations", len(results))
	if len(results) == 0 {
		return &CRMListResult{Results: "No organizations found."}, nil
	}
	var lines []string
	for _, org := range results {
		lines = append(lines, fmt.Sprintf("- %s (id=%d)", org.Name, org.ID))
	}
	return &CRMListResult{
		Results: fmt.Sprintf("Found %d organizations:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) listContacts(limit int) (*CRMListResult, error) {
	if t.contactService == nil {
		log.Printf("⚠️ ContactService not configured")
		return &CRMListResult{Results: "CRM service not configured. Unable to list contacts."}, nil
	}
	results, err := t.contactService.SearchContacts("", limit)
	if err != nil {
		log.Printf("❌ Contact list failed: %v", err)
		return &CRMListResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("📋 crm_list found %d contacts", len(results))
	if len(results) == 0 {
		return &CRMListResult{Results: "No contacts found."}, nil
	}
	var lines []string
	for _, c := range results {
		lines = append(lines, formatContact(c))
	}
	return &CRMListResult{
		Results: fmt.Sprintf("Found %d contacts:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) listAccounts(limit int) (*CRMListResult, error) {
	if t.accountService == nil {
		log.Printf("⚠️ AccountService not configured")
		return &CRMListResult{Results: "CRM service not configured. Unable to list accounts."}, nil
	}
	results, err := t.accountService.SearchAccounts("", nil, limit)
	if err != nil {
		log.Printf("❌ Account list failed: %v", err)
		return &CRMListResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("📋 crm_list found %d accounts", len(results))
	if len(results) == 0 {
		return &CRMListResult{Results: "No accounts found."}, nil
	}
	var lines []string
	for _, a := range results {
		lines = append(lines, fmt.Sprintf("- %s (type=%s, id=%d)", a.Name, a.Type, a.ID))
	}
	return &CRMListResult{
		Results: fmt.Sprintf("Found %d accounts:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) listJobs(limit int) (*CRMListResult, error) {
	if t.jobService == nil {
		log.Printf("⚠️ JobService not configured")
		return &CRMListResult{Results: "CRM service not configured. Unable to list jobs."}, nil
	}
	results, err := t.jobService.GetLeads(nil, nil)
	if err != nil {
		log.Printf("❌ Job list failed: %v", err)
		return &CRMListResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}
	log.Printf("📋 crm_list found %d jobs", len(results))
	if len(results) == 0 {
		return &CRMListResult{Results: "No jobs found."}, nil
	}
	if len(results) > limit {
		results = results[:limit]
	}
	var lines []string
	for _, j := range results {
		lines = append(lines, formatJobLead(j))
	}
	return &CRMListResult{
		Results: fmt.Sprintf("Found %d jobs:\n%s", len(results), strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func (t *CRMTools) listGeneric(entityType string, limit int) (*CRMListResult, error) {
	if t.entityLister == nil {
		log.Printf("⚠️ EntityLister not configured")
		return &CRMListResult{Results: "Generic entity listing not configured."}, nil
	}

	log.Printf("📋 crm_list generic for entity_type='%s', limit=%d", entityType, limit)

	results, err := t.entityLister.ListEntities(entityType, limit)
	if err != nil {
		log.Printf("❌ Generic list failed for %s: %v", entityType, err)
		return &CRMListResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}

	log.Printf("📋 crm_list found %d %s records", len(results), entityType)
	if len(results) == 0 {
		return &CRMListResult{Results: fmt.Sprintf("No %s records found.", entityType)}, nil
	}

	var lines []string
	for _, row := range results {
		lines = append(lines, formatGenericRow(row))
	}
	return &CRMListResult{
		Results: fmt.Sprintf("Found %d %s records:\n%s", len(results), entityType, strings.Join(lines, "\n")),
		Count:   len(results),
	}, nil
}

func formatGenericRow(row map[string]interface{}) string {
	id := row["id"]
	name := ""
	if n, ok := row["name"].(string); ok {
		name = n
	} else if t, ok := row["title"].(string); ok {
		name = t
	} else if d, ok := row["description"].(string); ok && len(d) > 50 {
		name = d[:50] + "..."
	} else if d, ok := row["description"].(string); ok {
		name = d
	}
	if name == "" {
		name = "unnamed"
	}
	return fmt.Sprintf("- %s (id=%v)", name, id)
}

func formatIndividual(ind serializers.IndividualBrief) string {
	name := strings.TrimSpace(fmt.Sprintf("%s %s", ind.FirstName, ind.LastName))
	email := "N/A"
	if ind.Email != nil {
		email = *ind.Email
	}
	return fmt.Sprintf("- %s (id=%d, email=%s)", name, ind.ID, email)
}

func formatContact(c serializers.ContactResponse) string {
	name := strings.TrimSpace(fmt.Sprintf("%s %s", c.FirstName, c.LastName))
	email := "N/A"
	if c.Email != nil {
		email = *c.Email
	}
	return fmt.Sprintf("- %s (id=%d, email=%s)", name, c.ID, email)
}

func formatJobLead(j serializers.JobDetailResponse) string {
	status := "unknown"
	if j.LeadStatus.Name != "" {
		status = j.LeadStatus.Name
	}
	return fmt.Sprintf("- %s at %s (status=%s, id=%d)", j.JobTitle, j.Account, status, j.ID)
}

// --- Propose CUD Operations ---

// CRMProposeCreateParams defines parameters for proposing entity creation
type CRMProposeCreateParams struct {
	EntityType string                 `json:"entity_type" jsonschema:"required,The type of CRM entity to create: contact, organization, individual, note, or task"`
	Data       map[string]interface{} `json:"data" jsonschema:"required,The entity data to create"`
}

// CRMProposeUpdateParams defines parameters for proposing entity update
type CRMProposeUpdateParams struct {
	EntityType string                 `json:"entity_type" jsonschema:"required,The type of CRM entity to update: contact, organization, individual, note, or task"`
	EntityID   int64                  `json:"entity_id" jsonschema:"required,The ID of the entity to update"`
	Data       map[string]interface{} `json:"data" jsonschema:"required,The fields to update"`
}

// CRMProposeDeleteParams defines parameters for proposing entity deletion
type CRMProposeDeleteParams struct {
	EntityType string `json:"entity_type" jsonschema:"required,The type of CRM entity to delete: contact, organization, individual, note, or task"`
	EntityID   int64  `json:"entity_id" jsonschema:"required,The ID of the entity to delete"`
}

// CRMProposeResult is the result of a propose operation
type CRMProposeResult struct {
	Success    bool   `json:"success"`
	ProposalID int64  `json:"proposal_id,omitempty"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// ProposeCreateCtx proposes creating a new CRM entity (queued for approval)
func (t *CRMTools) ProposeCreateCtx(ctx context.Context, params CRMProposeCreateParams) (*CRMProposeResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_create called for entity_type='%s'", entityType)

	if denied := t.checkPermission(ctx, entityType+":create"); denied != nil {
		return &CRMProposeResult{Success: false, Message: denied.Results}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeResult{Success: false, Message: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	payload, err := json.Marshal(params.Data)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: fmt.Sprintf("Failed to serialize data: %v", err)}, nil
	}

	result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, nil, "create", payload)
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

// ProposeUpdateCtx proposes updating a CRM entity (queued for approval)
func (t *CRMTools) ProposeUpdateCtx(ctx context.Context, params CRMProposeUpdateParams) (*CRMProposeResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_update called for entity_type='%s', entity_id=%d", entityType, params.EntityID)

	if denied := t.checkPermission(ctx, entityType+":update"); denied != nil {
		return &CRMProposeResult{Success: false, Message: denied.Results}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeResult{Success: false, Message: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	payload, err := json.Marshal(params.Data)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: fmt.Sprintf("Failed to serialize data: %v", err)}, nil
	}

	entityID := params.EntityID
	result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, &entityID, "update", payload)
	if err != nil {
		log.Printf("❌ ProposeUpdate failed: %v", err)
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	log.Printf("✅ ProposeUpdate succeeded: proposal_id=%d, status=%s", result.ProposalID, result.Status)
	return &CRMProposeResult{
		Success:    true,
		ProposalID: result.ProposalID,
		Status:     result.Status,
		Message:    result.Message,
	}, nil
}

// ProposeDeleteCtx proposes deleting a CRM entity (queued for approval)
func (t *CRMTools) ProposeDeleteCtx(ctx context.Context, params CRMProposeDeleteParams) (*CRMProposeResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("📝 crm_propose_delete called for entity_type='%s', entity_id=%d", entityType, params.EntityID)

	if denied := t.checkPermission(ctx, entityType+":delete"); denied != nil {
		return &CRMProposeResult{Success: false, Message: denied.Results}, nil
	}

	if t.proposalCreator == nil {
		return &CRMProposeResult{Success: false, Message: "Proposal service not configured"}, nil
	}

	tenantID, userID, err := t.getContextIDs(ctx)
	if err != nil {
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	entityID := params.EntityID
	result, err := t.proposalCreator.ProposeOperationBytes(tenantID, userID, entityType, &entityID, "delete", nil)
	if err != nil {
		log.Printf("❌ ProposeDelete failed: %v", err)
		return &CRMProposeResult{Success: false, Message: err.Error()}, nil
	}

	log.Printf("✅ ProposeDelete succeeded: proposal_id=%d, status=%s", result.ProposalID, result.Status)
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
