package repositories

import (
	"time"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"

	"agent/internal/models"
)

type EmbeddingRepository struct {
	db *gorm.DB
}

func NewEmbeddingRepository(db *gorm.DB) *EmbeddingRepository {
	return &EmbeddingRepository{db: db}
}

// EmbeddingChunk represents a chunk of text with its embedding vector
type EmbeddingChunk struct {
	ChunkIndex int
	ChunkText  string
	Embedding  pgvector.Vector
}

// EmbeddingDTO is a lightweight embedding result
type EmbeddingDTO struct {
	ID         int64
	SourceType string
	SourceID   int64
	ChunkIndex int
	ChunkText  string
	Embedding  pgvector.Vector
}

// SimilarResult holds a similarity search result
type SimilarResult struct {
	ID         int64
	SourceID   int64
	ChunkText  string
	Similarity float64
}

// UpsertDocumentChunks deletes old embeddings for a document and inserts new ones
func (r *EmbeddingRepository) UpsertDocumentChunks(tenantID, docID int64, chunks []EmbeddingChunk) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("source_type = ? AND source_id = ?", "document", docID).
			Delete(&models.Embedding{}).Error; err != nil {
			return err
		}

		embeddings := make([]models.Embedding, len(chunks))
		for i, c := range chunks {
			embeddings[i] = models.Embedding{
				TenantID:   tenantID,
				SourceType: "document",
				SourceID:   docID,
				ChunkIndex: c.ChunkIndex,
				ChunkText:  c.ChunkText,
				Embedding:  c.Embedding,
				CreatedAt:  time.Now(),
			}
		}
		return tx.Create(&embeddings).Error
	})
}

// GetDocumentEmbeddings returns embeddings for the given document IDs
func (r *EmbeddingRepository) GetDocumentEmbeddings(docIDs []int64) ([]EmbeddingDTO, error) {
	if len(docIDs) == 0 {
		return nil, nil
	}

	var rows []models.Embedding
	err := r.db.Where("source_type = ? AND source_id IN ?", "document", docIDs).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]EmbeddingDTO, len(rows))
	for i, row := range rows {
		result[i] = EmbeddingDTO{
			ID:         row.ID,
			SourceType: row.SourceType,
			SourceID:   row.SourceID,
			ChunkIndex: row.ChunkIndex,
			ChunkText:  row.ChunkText,
			Embedding:  row.Embedding,
		}
	}
	return result, nil
}

// InsertProposalEmbedding stores an embedding for an approved proposal
func (r *EmbeddingRepository) InsertProposalEmbedding(tenantID, proposalID, configID int64, text string, embedding pgvector.Vector) error {
	row := models.Embedding{
		TenantID:   tenantID,
		SourceType: "proposal",
		SourceID:   proposalID,
		ConfigID:   &configID,
		ChunkIndex: 0,
		ChunkText:  text,
		Embedding:  embedding,
		CreatedAt:  time.Now(),
	}
	return r.db.Create(&row).Error
}

// InsertEmbeddingWithType stores an embedding with a custom source_type
func (r *EmbeddingRepository) InsertEmbeddingWithType(tenantID, sourceID, configID int64, sourceType, text string, embedding pgvector.Vector) error {
	row := models.Embedding{
		TenantID:   tenantID,
		SourceType: sourceType,
		SourceID:   sourceID,
		ConfigID:   &configID,
		ChunkIndex: 0,
		ChunkText:  text,
		Embedding:  embedding,
		CreatedAt:  time.Now(),
	}
	return r.db.Create(&row).Error
}

// GetProposalCentroid returns the average embedding vector and count for a config's proposals (includes user approvals)
func (r *EmbeddingRepository) GetProposalCentroid(configID int64) (pgvector.Vector, int, error) {
	var result struct {
		Centroid pgvector.Vector
		Count    int
	}

	err := r.db.Raw(`SELECT AVG(embedding)::vector AS centroid, COUNT(*) AS count
		FROM "Embedding" WHERE config_id = ? AND source_type IN ('proposal', 'user_approved')`, configID).
		Scan(&result).Error
	if err != nil {
		return pgvector.Vector{}, 0, err
	}
	return result.Centroid, result.Count, nil
}

// GetRejectionCentroid returns the average embedding vector and count for a config's rejected proposals
func (r *EmbeddingRepository) GetRejectionCentroid(configID int64) (pgvector.Vector, int, error) {
	var result struct {
		Centroid pgvector.Vector
		Count    int
	}

	err := r.db.Raw(`SELECT AVG(embedding)::vector AS centroid, COUNT(*) AS count
		FROM "Embedding" WHERE config_id = ? AND source_type = 'user_rejected'`, configID).
		Scan(&result).Error
	if err != nil {
		return pgvector.Vector{}, 0, err
	}
	return result.Centroid, result.Count, nil
}

// DeleteBySource deletes all embeddings for a given source type and ID
func (r *EmbeddingRepository) DeleteBySource(sourceType string, sourceID int64) error {
	return r.db.Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Delete(&models.Embedding{}).Error
}

// FindSimilarDocChunks returns the top-N most similar document chunks to a query vector
func (r *EmbeddingRepository) FindSimilarDocChunks(docIDs []int64, queryVector pgvector.Vector, limit int) ([]SimilarResult, error) {
	if len(docIDs) == 0 {
		return nil, nil
	}

	var results []SimilarResult
	err := r.db.Raw(`SELECT id, source_id, chunk_text, 1 - (embedding <=> ?::vector) AS similarity
		FROM "Embedding"
		WHERE source_type = 'document' AND source_id IN ?
		ORDER BY embedding <=> ?::vector
		LIMIT ?`, queryVector, docIDs, queryVector, limit).
		Scan(&results).Error
	return results, err
}
