package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	pgvector "github.com/pgvector/pgvector-go"

	"agent/internal/repositories"
)

const (
	embeddingModel     = "voyage-4-large"
	embeddingDimension = 1024
	chunkMaxTokens     = 512
	chunkOverlapTokens = 64
	voyageEmbedURL     = "https://api.voyageai.com/v1/embeddings"
)

// DocumentEmbedder is the interface used by DocumentService
type DocumentEmbedder interface {
	EmbedDocumentChunks(ctx context.Context, tenantID, docID int64, content string) error
}

type EmbeddingService struct {
	repo       *repositories.EmbeddingRepository
	apiKey     string
	httpClient *http.Client
}

func NewEmbeddingService(repo *repositories.EmbeddingRepository, voyageAPIKey string) *EmbeddingService {
	return &EmbeddingService{
		repo:       repo,
		apiKey:     voyageAPIKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// EmbedTexts embeds a batch of texts via Voyage AI API
func (s *EmbeddingService) EmbedTexts(ctx context.Context, texts []string) ([]pgvector.Vector, error) {
	log.Printf("Embedding: calling Voyage AI %s for %d texts", embeddingModel, len(texts))
	rawVectors, err := s.callEmbeddingAPI(ctx, texts)
	if err != nil {
		return nil, err
	}

	vectors := make([]pgvector.Vector, len(rawVectors))
	for i, raw := range rawVectors {
		f32 := make([]float32, len(raw))
		for j, v := range raw {
			f32[j] = float32(v)
		}
		vectors[i] = pgvector.NewVector(f32)
	}
	return vectors, nil
}

// EmbedDocumentChunks chunks text, embeds, and stores for a document
func (s *EmbeddingService) EmbedDocumentChunks(ctx context.Context, tenantID, docID int64, content string) error {
	chunks := ChunkText(content, chunkMaxTokens, chunkOverlapTokens)
	if len(chunks) == 0 {
		return nil
	}

	log.Printf("📐 Embedding: chunked doc %d into %d chunks (%d chars total)", docID, len(chunks), len(content))

	vectors, err := s.EmbedTexts(ctx, chunks)
	if err != nil {
		return fmt.Errorf("embedding chunks: %w", err)
	}

	embChunks := make([]repositories.EmbeddingChunk, len(chunks))
	for i := range chunks {
		embChunks[i] = repositories.EmbeddingChunk{
			ChunkIndex: i,
			ChunkText:  chunks[i],
			Embedding:  vectors[i],
		}
	}

	if err := s.repo.UpsertDocumentChunks(tenantID, docID, embChunks); err != nil {
		return err
	}

	log.Printf("📐 Embedding: stored %d chunks for doc %d (tenant=%d)", len(embChunks), docID, tenantID)
	return nil
}

// EmbedProposal embeds a proposal's title+snippet and stores it
func (s *EmbeddingService) EmbedProposal(ctx context.Context, tenantID, proposalID, configID int64, title, snippet string) error {
	text := title + " - " + snippet
	vectors, err := s.EmbedTexts(ctx, []string{text})
	if err != nil {
		return fmt.Errorf("embedding proposal: %w", err)
	}

	if err := s.repo.InsertProposalEmbedding(tenantID, proposalID, configID, text, vectors[0]); err != nil {
		return err
	}

	log.Printf("📐 Embedding: stored proposal %d embedding (config=%d, %d chars)", proposalID, configID, len(text))
	return nil
}

// EmbedWithSourceType embeds text and stores it with a custom source_type
func (s *EmbeddingService) EmbedWithSourceType(ctx context.Context, tenantID, sourceID, configID int64, sourceType, text string) error {
	vectors, err := s.EmbedTexts(ctx, []string{text})
	if err != nil {
		return fmt.Errorf("embedding %s: %w", sourceType, err)
	}

	if err := s.repo.InsertEmbeddingWithType(tenantID, sourceID, configID, sourceType, text, vectors[0]); err != nil {
		return err
	}

	log.Printf("📐 Embedding: stored %s %d embedding (config=%d, %d chars)", sourceType, sourceID, configID, len(text))
	return nil
}

// DeleteBySource deletes all embeddings for a given source type and ID
func (s *EmbeddingService) DeleteBySource(sourceType string, sourceID int64) error {
	return s.repo.DeleteBySource(sourceType, sourceID)
}

// GetRejectionCentroid returns the rejection centroid vector and count for a config
func (s *EmbeddingService) GetRejectionCentroid(configID int64) (pgvector.Vector, int, error) {
	return s.repo.GetRejectionCentroid(configID)
}

// GetDocumentEmbeddings returns embeddings for the given document IDs
func (s *EmbeddingService) GetDocumentEmbeddings(docIDs []int64) ([]repositories.EmbeddingDTO, error) {
	return s.repo.GetDocumentEmbeddings(docIDs)
}

// GetProposalCentroid returns the centroid vector and count for a config's proposals
func (s *EmbeddingService) GetProposalCentroid(configID int64) (pgvector.Vector, int, error) {
	return s.repo.GetProposalCentroid(configID)
}

// FindSimilarDocChunks returns the top-N most similar document chunks to a query vector
func (s *EmbeddingService) FindSimilarDocChunks(docIDs []int64, queryVector pgvector.Vector, limit int) ([]repositories.SimilarResult, error) {
	return s.repo.FindSimilarDocChunks(docIDs, queryVector, limit)
}

// ChunkText splits text into chunks at sentence boundaries
func ChunkText(text string, maxTokens, overlapTokens int) []string {
	// Approximate: 1 token ≈ 4 chars
	maxChars := maxTokens * 4
	overlapChars := overlapTokens * 4

	text = strings.TrimSpace(text)
	if len(text) <= maxChars {
		return []string{text}
	}

	sentences := splitSentences(text)
	var chunks []string
	var current strings.Builder
	var overlapBuf []string

	for _, sent := range sentences {
		if current.Len()+len(sent) > maxChars && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			// Build overlap from recent sentences
			current.Reset()
			for _, ob := range overlapBuf {
				current.WriteString(ob)
			}
		}

		current.WriteString(sent)

		// Track recent sentences for overlap
		overlapBuf = append(overlapBuf, sent)
		for totalLen(overlapBuf) > overlapChars && len(overlapBuf) > 1 {
			overlapBuf = overlapBuf[1:]
		}
	}

	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}
	return chunks
}

func totalLen(ss []string) int {
	n := 0
	for _, s := range ss {
		n += len(s)
	}
	return n
}

func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	for _, r := range text {
		current.WriteRune(r)
		if (r == '.' || r == '!' || r == '?') && current.Len() > 10 {
			sentences = append(sentences, current.String())
			current.Reset()
		}
		if r == '\n' && current.Len() > 0 {
			s := current.String()
			trimmed := strings.TrimSpace(s)
			if len(trimmed) > 0 && !isAllSpace(s) {
				sentences = append(sentences, s)
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		sentences = append(sentences, current.String())
	}
	return sentences
}

func isAllSpace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// Voyage AI embeddings API types (OpenAI-compatible shape)

type embeddingRequest struct {
	Model           string   `json:"model"`
	Input           []string `json:"input"`
	InputType       string   `json:"input_type,omitempty"`
	OutputDimension int      `json:"output_dimension,omitempty"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *EmbeddingService) callEmbeddingAPI(ctx context.Context, texts []string) ([][]float64, error) {
	body, err := json.Marshal(embeddingRequest{
		Model:           embeddingModel,
		Input:           texts,
		InputType:       "document",
		OutputDimension: embeddingDimension,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, voyageEmbedURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Voyage AI embeddings API error (status %d): %s", resp.StatusCode, string(respBody[:min(len(respBody), 200)]))
		return nil, fmt.Errorf("voyage api error: status %d", resp.StatusCode)
	}

	var result embeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("voyage api: %s", result.Error.Message)
	}

	vectors := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		vectors[i] = d.Embedding
	}
	return vectors, nil
}
