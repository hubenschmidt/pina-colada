package filtering

import (
	"context"

	pgvector "github.com/pgvector/pgvector-go"

	"agent/internal/repositories"
)

// EmbeddingProvider abstracts embedding operations for vector pre-filtering
type EmbeddingProvider interface {
	EmbedTexts(ctx context.Context, texts []string) ([]pgvector.Vector, error)
	GetDocumentEmbeddings(docIDs []int64) ([]repositories.EmbeddingDTO, error)
	GetRejectionCentroid(configID int64) (pgvector.Vector, int, error)
	EmbedDocumentChunks(ctx context.Context, tenantID, docID int64, content string) error
}

// DocumentLoader abstracts document retrieval for context building
type DocumentLoader interface {
	GetDocumentByID(id int64) (*DocumentResult, error)
}

// DocumentResult wraps a document with its content bytes
type DocumentResult struct {
	Document DocumentMeta
	Content  []byte
}

// DocumentMeta holds document metadata
type DocumentMeta struct {
	Filename    string
	ContentType string
}
