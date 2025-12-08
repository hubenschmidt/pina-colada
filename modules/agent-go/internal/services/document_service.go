package services

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/storage"
)

const MaxFileSize = 10 * 1024 * 1024 // 10MB

var entityTypeMap = map[string]string{
	"project":      "Project",
	"lead":         "Lead",
	"account":      "Account",
	"organization": "Organization",
	"individual":   "Individual",
	"contact":      "Contact",
}

func normalizeEntityType(entityType string) string {
	if mapped, ok := entityTypeMap[strings.ToLower(entityType)]; ok {
		return mapped
	}
	return entityType
}

type DocumentService struct {
	docRepo *repositories.DocumentRepository
	storage storage.Backend
}

func NewDocumentService(docRepo *repositories.DocumentRepository, storageBackend storage.Backend) *DocumentService {
	return &DocumentService{docRepo: docRepo, storage: storageBackend}
}

func (s *DocumentService) GetDocumentsByEntity(entityType string, entityID int64, tenantID *int64, page, pageSize int, orderBy, order string) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	result, err := s.docRepo.FindByEntity(entityType, entityID, tenantID, params)
	if err != nil {
		return nil, err
	}

	resp := serializers.NewPagedResponse(result.Items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// EntityLinkInput holds data for linking a document to an entity
type EntityLinkInput struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
}

// LinkDocument links a document to an entity
func (s *DocumentService) LinkDocument(documentID int64, input EntityLinkInput) error {
	return s.docRepo.LinkToEntity(documentID, input.EntityType, input.EntityID)
}

// UnlinkDocument removes a document's link to an entity
func (s *DocumentService) UnlinkDocument(documentID int64, entityType string, entityID int64) error {
	normalizedType := normalizeEntityType(entityType)
	return s.docRepo.UnlinkFromEntity(documentID, normalizedType, entityID)
}

// GetDocumentByID returns a single document by ID
func (s *DocumentService) GetDocumentByID(id int64) (*repositories.DocumentDTO, error) {
	return s.docRepo.FindDocumentByID(id)
}

func (s *DocumentService) GetDocuments(entityType *string, entityID *int64, tenantID *int64, search string, page, pageSize int, orderBy, order string) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search
	result, err := s.docRepo.FindAll(entityType, entityID, tenantID, params)
	if err != nil {
		return nil, err
	}

	resp := serializers.NewPagedResponse(result.Items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// UploadDocumentInput holds data for uploading a document
type UploadDocumentInput struct {
	TenantID    int64
	UserID      int64
	Filename    string
	ContentType string
	Data        io.Reader
	Size        int64
	Description *string
	EntityType  *string
	EntityID    *int64
}

// UploadDocument uploads a file to storage and creates DB records
func (s *DocumentService) UploadDocument(input UploadDocumentInput) (*repositories.DocumentDTO, error) {
	if input.Size > MaxFileSize {
		return nil, fmt.Errorf("file exceeds 10MB limit")
	}

	// Generate storage path: {tenant_id}/{timestamp}/{filename}
	timestamp := time.Now().UnixMilli()
	storagePath := fmt.Sprintf("%d/%d/%s", input.TenantID, timestamp, input.Filename)

	// Upload to storage
	if err := s.storage.Upload(storagePath, input.Data, input.ContentType, input.Size); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create database records
	doc, err := s.docRepo.CreateDocument(repositories.CreateDocumentInput{
		TenantID:    input.TenantID,
		UserID:      input.UserID,
		Filename:    input.Filename,
		ContentType: input.ContentType,
		StoragePath: storagePath,
		FileSize:    input.Size,
		Description: input.Description,
	})
	if err != nil {
		// Try to clean up the uploaded file
		s.storage.Delete(storagePath)
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// Link to entity if provided
	if input.EntityType != nil && input.EntityID != nil {
		normalizedType := normalizeEntityType(*input.EntityType)
		if err := s.docRepo.LinkToEntity(doc.ID, normalizedType, *input.EntityID); err != nil {
			// Don't fail the whole upload, just log or ignore
		}
	}

	return doc, nil
}
