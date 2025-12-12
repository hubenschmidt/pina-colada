package tools

import (
	"fmt"
	"log"
	"strings"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/services"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// DocumentTools holds document-related tools for the agent
type DocumentTools struct {
	docService *services.DocumentService
}

// NewDocumentTools creates document tools with service dependencies
func NewDocumentTools(docService *services.DocumentService) *DocumentTools {
	return &DocumentTools{docService: docService}
}

// --- Tool Parameter Structs ---

// SearchEntityDocumentsParams defines parameters for searching documents linked to an entity
type SearchEntityDocumentsParams struct {
	EntityType string `json:"entity_type" jsonschema:"The type of entity (individual, organization, account, contact)"`
	EntityID   int64  `json:"entity_id" jsonschema:"The ID of the entity"`
}

// SearchEntityDocumentsResult is the result of searching entity documents
type SearchEntityDocumentsResult struct {
	Results string `json:"results"`
	Count   int    `json:"count"`
}

// ReadDocumentParams defines parameters for reading a document
type ReadDocumentParams struct {
	DocumentID int64 `json:"document_id" jsonschema:"The ID of the document to read"`
}

// ReadDocumentResult is the result of reading a document
type ReadDocumentResult struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
}

// --- Tool Functions ---

// SearchEntityDocuments searches for documents linked to a specific entity
func (t *DocumentTools) SearchEntityDocuments(ctx tool.Context, params SearchEntityDocumentsParams) (*SearchEntityDocumentsResult, error) {
	log.Printf("📄 search_entity_documents called with entity_type='%s', entity_id=%d", params.EntityType, params.EntityID)

	if t.docService == nil {
		log.Printf("⚠️ DocumentService not configured")
		return &SearchEntityDocumentsResult{Results: "Document service not configured. Unable to search documents."}, nil
	}

	// Normalize entity type to Title case (e.g., "individual" -> "Individual")
	entityType := strings.Title(strings.ToLower(params.EntityType))
	resp, err := t.docService.GetDocumentsByEntity(entityType, params.EntityID, nil, 1, 20, "updated_at", "DESC")
	if err != nil {
		log.Printf("❌ Document search failed: %v", err)
		return &SearchEntityDocumentsResult{Results: fmt.Sprintf("Error: %v", err)}, nil
	}

	// Handle different item types from the paged response
	var lines []string
	switch items := resp.Items.(type) {
	case []repositories.DocumentDTO:
		if len(items) == 0 {
			return &SearchEntityDocumentsResult{
				Results: fmt.Sprintf("No documents found for %s id=%d", params.EntityType, params.EntityID),
				Count:   0,
			}, nil
		}
		for _, doc := range items {
			desc := ""
			if doc.Description != nil && *doc.Description != "" {
				desc = fmt.Sprintf(" - %s", *doc.Description)
			}
			lines = append(lines, fmt.Sprintf("- id=%d: %s%s (%s)", doc.ID, doc.Filename, desc, doc.ContentType))
		}
	case []interface{}:
		if len(items) == 0 {
			return &SearchEntityDocumentsResult{
				Results: fmt.Sprintf("No documents found for %s id=%d", params.EntityType, params.EntityID),
				Count:   0,
			}, nil
		}
		for _, item := range items {
			if doc, ok := item.(map[string]interface{}); ok {
				id := doc["id"]
				filename := doc["filename"]
				contentType := doc["content_type"]
				desc := ""
				if d, ok := doc["description"].(string); ok && d != "" {
					desc = fmt.Sprintf(" - %s", d)
				}
				lines = append(lines, fmt.Sprintf("- id=%v: %v%s (%v)", id, filename, desc, contentType))
			}
		}
	default:
		log.Printf("⚠️ Unexpected items type: %T", resp.Items)
		return &SearchEntityDocumentsResult{
			Results: fmt.Sprintf("No documents found for %s id=%d", params.EntityType, params.EntityID),
			Count:   0,
		}, nil
	}

	log.Printf("📄 search_entity_documents found %d documents", len(lines))
	return &SearchEntityDocumentsResult{
		Results: fmt.Sprintf("Found %d documents for %s id=%d:\n%s", len(lines), params.EntityType, params.EntityID, strings.Join(lines, "\n")),
		Count:   len(lines),
	}, nil
}

// ReadDocument reads the content of a document by ID
func (t *DocumentTools) ReadDocument(ctx tool.Context, params ReadDocumentParams) (*ReadDocumentResult, error) {
	log.Printf("📖 read_document called with document_id=%d", params.DocumentID)

	if t.docService == nil {
		log.Printf("⚠️ DocumentService not configured")
		return &ReadDocumentResult{Content: "Document service not configured. Unable to read document."}, nil
	}

	// Get document metadata
	doc, err := t.docService.GetDocumentByID(params.DocumentID)
	if err != nil {
		log.Printf("❌ Document lookup failed: %v", err)
		return &ReadDocumentResult{Content: fmt.Sprintf("Error: %v", err)}, nil
	}
	if doc == nil {
		return &ReadDocumentResult{Content: fmt.Sprintf("Document %d not found", params.DocumentID)}, nil
	}

	// Download content (using tenant ID 1 as default for now - should be passed via context)
	result, err := t.docService.DownloadDocument(params.DocumentID, doc.TenantID)
	if err != nil {
		log.Printf("❌ Document download failed: %v", err)
		return &ReadDocumentResult{Content: fmt.Sprintf("Error downloading document: %v", err)}, nil
	}

	var content string
	if result.Content != nil {
		// For text files, return content directly
		if strings.Contains(doc.ContentType, "text/") || strings.Contains(doc.ContentType, "application/json") {
			content = string(result.Content)
		} else if strings.Contains(doc.ContentType, "pdf") {
			// TODO: Add PDF text extraction
			content = "[PDF content - text extraction not yet implemented]"
		} else {
			content = fmt.Sprintf("[Binary content - %s, %d bytes]", doc.ContentType, len(result.Content))
		}
	} else if result.RedirectURL != nil {
		content = fmt.Sprintf("[Document available at: %s]", *result.RedirectURL)
	} else {
		content = "Unable to retrieve document content"
	}

	// Truncate for token optimization
	const maxChars = 15000
	if len(content) > maxChars {
		content = content[:maxChars] + fmt.Sprintf("\n\n[Truncated - %d total chars]", len(content))
	}

	log.Printf("📖 read_document returning content for '%s' (%d chars)", doc.Filename, len(content))
	return &ReadDocumentResult{
		Content:  fmt.Sprintf("=== %s ===\n\n%s", doc.Filename, content),
		Filename: doc.Filename,
	}, nil
}

// BuildTools returns ADK tool.Tool instances for document operations
func (t *DocumentTools) BuildTools() ([]tool.Tool, error) {
	searchTool, err := functiontool.New(
		functiontool.Config{
			Name:        "search_entity_documents",
			Description: "Search for documents linked to a specific entity (individual, organization, account, contact). Use this to find resumes, cover letters, or other documents associated with a CRM record.",
		},
		t.SearchEntityDocuments,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create search_entity_documents tool: %w", err)
	}

	readTool, err := functiontool.New(
		functiontool.Config{
			Name:        "read_document",
			Description: "Read the content of a document by its ID. Use this after search_entity_documents to read a specific document like a resume.",
		},
		t.ReadDocument,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create read_document tool: %w", err)
	}

	return []tool.Tool{searchTool, readTool}, nil
}
