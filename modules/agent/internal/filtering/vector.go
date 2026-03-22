package filtering

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	pgvector "github.com/pgvector/pgvector-go"

	"agent/internal/repositories"
)

// VectorPreFilterConfig holds config values needed for vector pre-filtering
type VectorPreFilterConfig struct {
	ConfigID            int64
	TenantID            int64
	SourceDocumentIDs   []int64
	SimilarityThreshold float64
}

// VectorPreFilter uses vector similarity to reduce results before LLM review
func VectorPreFilter(results []JobResult, cfg VectorPreFilterConfig, embeddingService EmbeddingProvider, docLoader DocumentLoader) []JobResult {
	if len(results) == 0 || len(cfg.SourceDocumentIDs) == 0 {
		return results
	}

	docEmbeddings, err := embeddingService.GetDocumentEmbeddings(cfg.SourceDocumentIDs)
	if err != nil {
		log.Printf("Filtering: vector pre-filter skipped (embedding lookup error): %v", err)
		return results
	}
	if len(docEmbeddings) == 0 {
		docEmbeddings = BackfillDocumentEmbeddings(cfg.TenantID, cfg.SourceDocumentIDs, embeddingService, docLoader)
	}
	if len(docEmbeddings) == 0 {
		log.Printf("Filtering: vector pre-filter skipped (no doc embeddings after backfill)")
		return results
	}

	texts := make([]string, len(results))
	for i, r := range results {
		texts[i] = EmbeddingText(r)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vectors, err := embeddingService.EmbedTexts(ctx, texts)
	if err != nil {
		log.Printf("Filtering: vector pre-filter embed failed: %v", err)
		return results
	}

	var rejectionCentroid pgvector.Vector
	var hasRejectionCentroid bool
	rejCentroid, rejCount, rejErr := embeddingService.GetRejectionCentroid(cfg.ConfigID)
	if rejErr == nil && rejCount >= 3 {
		rejectionCentroid = rejCentroid
		hasRejectionCentroid = true
		log.Printf("Filtering: rejection centroid loaded (%d user rejections)", rejCount)
	}

	type scored struct {
		job   JobResult
		score float64
	}
	var scoredResults []scored
	for i, vec := range vectors {
		positiveScore := MaxCosineSimilarity(vec, docEmbeddings)
		finalScore := positiveScore
		if hasRejectionCentroid {
			rejSim := CosineSimilarity(vec.Slice(), rejectionCentroid.Slice())
			finalScore = positiveScore - (rejSim * 0.3)
			log.Printf("  vector score [%d] %.3f (pos=%.3f, rej=%.3f) %s — %s", i, finalScore, positiveScore, rejSim, results[i].Title, passOrFail(finalScore, cfg.SimilarityThreshold))
		} else {
			log.Printf("  vector score [%d] %.3f %s — %s", i, finalScore, results[i].Title, passOrFail(finalScore, cfg.SimilarityThreshold))
		}
		if finalScore >= cfg.SimilarityThreshold {
			scoredResults = append(scoredResults, scored{job: results[i], score: finalScore})
		}
	}

	// Sort descending by score
	for i := 0; i < len(scoredResults)-1; i++ {
		for j := i + 1; j < len(scoredResults); j++ {
			if scoredResults[j].score > scoredResults[i].score {
				scoredResults[i], scoredResults[j] = scoredResults[j], scoredResults[i]
			}
		}
	}

	filtered := make([]JobResult, len(scoredResults))
	for i, sr := range scoredResults {
		filtered[i] = sr.job
	}

	log.Printf("Filtering: vector pre-filter: %d/%d results passed (threshold=%.2f)", len(filtered), len(results), cfg.SimilarityThreshold)
	return filtered
}

// BackfillDocumentEmbeddings lazily embeds source documents that have no stored embeddings yet.
func BackfillDocumentEmbeddings(tenantID int64, docIDs []int64, embeddingService EmbeddingProvider, docLoader DocumentLoader) []repositories.EmbeddingDTO {
	if docLoader == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, docID := range docIDs {
		result, err := docLoader.GetDocumentByID(docID)
		if err != nil || result == nil || result.Content == nil {
			log.Printf("Filtering: backfill skip doc %d (fetch failed): %v", docID, err)
			continue
		}
		content := ExtractDocumentContent(result.Document.ContentType, result.Content)
		if content == "" {
			log.Printf("Filtering: backfill skip doc %d (no extractable text)", docID)
			continue
		}
		log.Printf("Filtering: backfilling embeddings for doc %d (%d chars)", docID, len(content))
		if err := embeddingService.EmbedDocumentChunks(ctx, tenantID, docID, content); err != nil {
			log.Printf("Filtering: backfill embed failed for doc %d: %v", docID, err)
		}
	}

	embeddings, _ := embeddingService.GetDocumentEmbeddings(docIDs)
	return embeddings
}

// MaxCosineSimilarity computes the max cosine similarity between a vector and document embeddings
func MaxCosineSimilarity(vec pgvector.Vector, docEmbeddings []repositories.EmbeddingDTO) float64 {
	maxSim := -1.0
	vecSlice := vec.Slice()
	for _, de := range docEmbeddings {
		sim := CosineSimilarity(vecSlice, de.Embedding.Slice())
		if sim > maxSim {
			maxSim = sim
		}
	}
	return maxSim
}

// CosineSimilarity computes cosine similarity between two float32 slices
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// EmbeddingText returns the text to embed for a job result
func EmbeddingText(r JobResult) string {
	if r.FullText != "" {
		return r.FullText
	}
	return r.Title + " - " + r.Snippet
}

// TruncatedJobDescription truncates a job description for display
func TruncatedJobDescription(fullText string, maxLen int) string {
	if fullText == "" {
		return ""
	}
	if len(fullText) > maxLen {
		fullText = fullText[:maxLen] + "..."
	}
	return fmt.Sprintf("Job Description:\n%s\n", fullText)
}

func passOrFail(score, threshold float64) string {
	if score >= threshold {
		return "PASS"
	}
	return "FAIL"
}
