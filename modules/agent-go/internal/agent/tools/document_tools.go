package tools

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/services"
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

// --- Helper Functions ---

// extractPDFText extracts text content from PDF bytes
func extractPDFText(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}

	var text strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		extractPageText(reader.Page(i), i, &text)
	}

	return text.String(), nil
}

func extractPageText(page pdf.Page, pageNum int, text *strings.Builder) {
	if page.V.IsNull() {
		return
	}
	pageText, err := page.GetPlainText(nil)
	if err != nil {
		log.Printf("⚠️ Failed to extract text from page %d: %v", pageNum, err)
		return
	}
	text.WriteString(pageText)
	text.WriteString("\n")
}

// extractContent extracts readable content from a document based on its type
func extractContent(doc *schemas.DocumentResponse, result *services.DownloadDocumentResult) string {
	const maxChars = 15000

	if result.Content == nil && result.RedirectURL != nil {
		return fmt.Sprintf("[Document available at: %s]", *result.RedirectURL)
	}
	if result.Content == nil {
		return "Unable to retrieve document content"
	}

	content := extractByContentType(doc.ContentType, result.Content)

	if len(content) > maxChars {
		return content[:maxChars] + fmt.Sprintf("\n\n[Truncated - %d total chars]", len(content))
	}
	return content
}

// extractByContentType returns content string based on MIME type
func extractByContentType(contentType string, data []byte) string {
	if strings.Contains(contentType, "text/") || strings.Contains(contentType, "application/json") {
		return string(data)
	}

	if strings.Contains(contentType, "pdf") {
		return extractPDFContent(data)
	}

	return fmt.Sprintf("[Binary content - %s, %d bytes]", contentType, len(data))
}

// extractPDFContent extracts text from PDF data with error handling
func extractPDFContent(data []byte) string {
	pdfText, err := extractPDFText(data)
	if err != nil {
		log.Printf("⚠️ PDF extraction failed: %v", err)
		return fmt.Sprintf("[PDF extraction failed: %v]", err)
	}

	if strings.TrimSpace(pdfText) == "" {
		return "[PDF contains no extractable text - may be image-based]"
	}

	return pdfText
}

// --- Tool Functions ---

// SearchEntityDocumentsCtx searches for documents linked to a specific entity.
func (t *DocumentTools) SearchEntityDocumentsCtx(ctx context.Context, params SearchEntityDocumentsParams) (*SearchEntityDocumentsResult, error) {
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
	lines := formatDocumentItems(resp.Items)
	if lines == nil {
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

func formatDocumentItems(items interface{}) []string {
	if docs, ok := items.([]schemas.DocumentResponse); ok {
		return formatDocuments(docs)
	}
	if rawItems, ok := items.([]interface{}); ok {
		return formatRawDocumentItems(rawItems)
	}
	log.Printf("⚠️ Unexpected items type: %T", items)
	return nil
}

func formatDocuments(docs []schemas.DocumentResponse) []string {
	if len(docs) == 0 {
		return nil
	}
	var lines []string
	for _, doc := range docs {
		desc := ""
		if doc.Description != nil && *doc.Description != "" {
			desc = fmt.Sprintf(" - %s", *doc.Description)
		}
		lines = append(lines, fmt.Sprintf("- id=%d: %s%s (%s)", doc.ID, doc.Filename, desc, doc.ContentType))
	}
	return lines
}

func formatRawDocumentItems(items []interface{}) []string {
	if len(items) == 0 {
		return nil
	}
	var lines []string
	for _, item := range items {
		doc, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("⚠️ Unexpected item type in slice: %T", item)
			return nil
		}
		id := doc["id"]
		filename := doc["filename"]
		contentType := doc["content_type"]
		desc := ""
		if d, ok := doc["description"].(string); ok && d != "" {
			desc = fmt.Sprintf(" - %s", d)
		}
		lines = append(lines, fmt.Sprintf("- id=%v: %v%s (%v)", id, filename, desc, contentType))
	}
	return lines
}

// ReadDocumentCtx reads the content of a document by ID.
func (t *DocumentTools) ReadDocumentCtx(ctx context.Context, params ReadDocumentParams) (*ReadDocumentResult, error) {
	log.Printf("📖 read_document called with document_id=%d", params.DocumentID)

	if t.docService == nil {
		log.Printf("⚠️ DocumentService not configured")
		return &ReadDocumentResult{Content: "Document service not configured. Unable to read document."}, nil
	}

	doc, err := t.docService.GetDocumentByID(params.DocumentID)
	if err != nil {
		log.Printf("❌ Document lookup failed: %v", err)
		return &ReadDocumentResult{Content: fmt.Sprintf("Error: %v", err)}, nil
	}
	if doc == nil {
		return &ReadDocumentResult{Content: fmt.Sprintf("Document %d not found", params.DocumentID)}, nil
	}

	result, err := t.docService.DownloadDocument(params.DocumentID, doc.TenantID)
	if err != nil {
		log.Printf("❌ Document download failed: %v", err)
		return &ReadDocumentResult{Content: fmt.Sprintf("Error downloading document: %v", err)}, nil
	}

	content := extractContent(doc, result)

	log.Printf("📖 read_document returning content for '%s' (%d chars)", doc.Filename, len(content))
	return &ReadDocumentResult{
		Content:  fmt.Sprintf("=== %s ===\n\n%s", doc.Filename, content),
		Filename: doc.Filename,
	}, nil
}

