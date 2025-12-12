package tools

import (
	"fmt"
	"log"
	"strings"

	"github.com/pina-colada-co/agent-go/internal/services"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// CRMTools holds CRM-related tools for the agent
type CRMTools struct {
	indService *services.IndividualService
	orgService *services.OrganizationService
}

// NewCRMTools creates CRM tools with service dependencies
func NewCRMTools(indService *services.IndividualService, orgService *services.OrganizationService) *CRMTools {
	return &CRMTools{
		indService: indService,
		orgService: orgService,
	}
}

// --- Tool Parameter Structs ---

// CRMLookupParams defines parameters for CRM entity lookup
type CRMLookupParams struct {
	EntityType string `json:"entity_type" jsonschema:"The type of CRM entity to search: individual, organization, account, or contact"`
	Query      string `json:"query" jsonschema:"Search term for entity lookup (name, email, etc.)"`
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
	EntityType string `json:"entity_type" jsonschema:"The type of CRM entity to list: individual, organization, account, or contact"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of results to return (default 20)"`
}

// CRMListResult is the result of a list operation
type CRMListResult struct {
	Results string `json:"results"`
	Count   int    `json:"count"`
}

// --- Tool Functions ---

// Lookup searches for CRM entities by type and query
func (t *CRMTools) Lookup(ctx tool.Context, params CRMLookupParams) (*CRMLookupResult, error) {
	entityType := strings.ToLower(params.EntityType)
	log.Printf("🔍 crm_lookup called with entity_type='%s', query='%s'", entityType, params.Query)

	switch entityType {
	case "individual":
		if t.indService == nil {
			log.Printf("⚠️ IndividualService not configured")
			return &CRMLookupResult{Results: "CRM service not configured. Unable to look up individuals."}, nil
		}
		results, err := t.indService.SearchIndividuals(params.Query, nil, 10)
		if err != nil {
			log.Printf("❌ Individual lookup failed: %v", err)
			return &CRMLookupResult{Results: fmt.Sprintf("Error: %v", err)}, nil
		}
		log.Printf("🔍 crm_lookup found %d individuals", len(results))
		if len(results) == 0 {
			return &CRMLookupResult{Results: fmt.Sprintf("No individuals found matching '%s'", params.Query)}, nil
		}
		var lines []string
		for _, ind := range results {
			name := strings.TrimSpace(fmt.Sprintf("%s %s", ind.FirstName, ind.LastName))
			email := "N/A"
			if ind.Email != nil {
				email = *ind.Email
			}
			lines = append(lines, fmt.Sprintf("- %s (id=%d, email=%s)", name, ind.ID, email))
		}
		return &CRMLookupResult{
			Results: fmt.Sprintf("Found %d individuals:\n%s", len(results), strings.Join(lines, "\n")),
			Count:   len(results),
		}, nil

	case "organization":
		if t.orgService == nil {
			log.Printf("⚠️ OrganizationService not configured")
			return &CRMLookupResult{Results: "CRM service not configured. Unable to look up organizations."}, nil
		}
		results, err := t.orgService.SearchOrganizations(params.Query, nil, 10)
		if err != nil {
			log.Printf("❌ Organization lookup failed: %v", err)
			return &CRMLookupResult{Results: fmt.Sprintf("Error: %v", err)}, nil
		}
		log.Printf("🔍 crm_lookup found %d organizations", len(results))
		if len(results) == 0 {
			return &CRMLookupResult{Results: fmt.Sprintf("No organizations found matching '%s'", params.Query)}, nil
		}
		var lines []string
		for _, org := range results {
			lines = append(lines, fmt.Sprintf("- %s (id=%d)", org.Name, org.ID))
		}
		return &CRMLookupResult{
			Results: fmt.Sprintf("Found %d organizations:\n%s", len(results), strings.Join(lines, "\n")),
			Count:   len(results),
		}, nil

	default:
		log.Printf("⚠️  Unknown entity type: %s", entityType)
		return &CRMLookupResult{
			Results: fmt.Sprintf("Unknown entity type: %s. Supported: individual, organization", entityType),
		}, nil
	}
}

// List returns entities of a given type
func (t *CRMTools) List(ctx tool.Context, params CRMListParams) (*CRMListResult, error) {
	entityType := strings.ToLower(params.EntityType)
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	log.Printf("📋 crm_list called with entity_type='%s', limit=%d", entityType, limit)

	switch entityType {
	case "individual":
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
			name := strings.TrimSpace(fmt.Sprintf("%s %s", ind.FirstName, ind.LastName))
			email := "N/A"
			if ind.Email != nil {
				email = *ind.Email
			}
			lines = append(lines, fmt.Sprintf("- %s (id=%d, email=%s)", name, ind.ID, email))
		}
		return &CRMListResult{
			Results: fmt.Sprintf("Found %d individuals:\n%s", len(results), strings.Join(lines, "\n")),
			Count:   len(results),
		}, nil

	case "organization":
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

	default:
		log.Printf("⚠️  Unknown entity type: %s", entityType)
		return &CRMListResult{
			Results: fmt.Sprintf("Unknown entity type: %s. Supported: individual, organization", entityType),
		}, nil
	}
}

// BuildTools returns ADK tool.Tool instances for CRM operations
func (t *CRMTools) BuildTools() ([]tool.Tool, error) {
	lookupTool, err := functiontool.New(
		functiontool.Config{
			Name:        "crm_lookup",
			Description: "Search for CRM entities (individuals, organizations) by name, email, or other attributes. Use this to find specific contacts or companies.",
		},
		t.Lookup,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lookup tool: %w", err)
	}

	listTool, err := functiontool.New(
		functiontool.Config{
			Name:        "crm_list",
			Description: "List CRM entities of a specific type. Use this to see all individuals or organizations in the system.",
		},
		t.List,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create list tool: %w", err)
	}

	return []tool.Tool{lookupTool, listTool}, nil
}
