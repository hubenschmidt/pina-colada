package filtering

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ledongthuc/pdf"
)

// LoadDocumentContext loads and formats document content for LLM context
func LoadDocumentContext(docIDs []int64, docLoader DocumentLoader) string {
	if docLoader == nil || len(docIDs) == 0 {
		return ""
	}

	const maxTotalChars = 15000

	results := make([]string, len(docIDs))
	var wg sync.WaitGroup

	for i, docID := range docIDs {
		wg.Add(1)
		go func(idx int, id int64) {
			defer wg.Done()
			results[idx] = fetchDocumentContent(id, docLoader)
		}(i, docID)
	}
	wg.Wait()

	var builder strings.Builder
	for _, content := range results {
		if content != "" {
			builder.WriteString(content)
		}
	}

	return truncateDocumentContext(builder.String(), maxTotalChars, len(docIDs))
}

func fetchDocumentContent(docID int64, docLoader DocumentLoader) string {
	result, err := docLoader.GetDocumentByID(docID)
	if err != nil {
		log.Printf("Filtering: failed to load doc %d: %v", docID, err)
		return ""
	}
	if result == nil || result.Content == nil {
		return ""
	}

	content := ExtractDocumentContent(result.Document.ContentType, result.Content)
	if content == "" {
		return ""
	}

	return fmt.Sprintf("=== Document: %s ===\n%s\n\n", result.Document.Filename, content)
}

func truncateDocumentContext(result string, maxChars, docCount int) string {
	if len(result) > maxChars {
		result = result[:maxChars] + "\n[Truncated]"
	}
	log.Printf("Filtering: loaded %d documents (%d chars)", docCount, len(result))
	return result
}

// ExtractDocumentContent extracts text from document bytes based on content type
func ExtractDocumentContent(contentType string, data []byte) string {
	if strings.Contains(contentType, "text/") || strings.Contains(contentType, "application/json") {
		return string(data)
	}
	if strings.Contains(contentType, "pdf") {
		return extractPDFTextFromBytes(data)
	}
	return ""
}

func extractPDFTextFromBytes(data []byte) string {
	reader := bytes.NewReader(data)
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return ""
	}

	var text strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		appendPDFPageText(&text, pdfReader.Page(i))
	}
	return text.String()
}

func appendPDFPageText(text *strings.Builder, page pdf.Page) {
	if page.V.IsNull() {
		return
	}
	pageText, err := page.GetPlainText(nil)
	if err != nil {
		return
	}
	text.WriteString(pageText)
	text.WriteString("\n")
}

// ExtractDocumentNames extracts document names from formatted context
func ExtractDocumentNames(context string) []string {
	var names []string
	lines := strings.Split(context, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "=== Document: ") && strings.HasSuffix(line, " ===") {
			name := strings.TrimPrefix(line, "=== Document: ")
			name = strings.TrimSuffix(name, " ===")
			names = append(names, name)
		}
	}
	return names
}

// TruncateString truncates a string to maxLen chars with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
