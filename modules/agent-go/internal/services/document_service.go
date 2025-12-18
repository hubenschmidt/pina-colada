package services

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

const MaxFileSize = 10 * 1024 * 1024 // 10MB

// ErrDocumentNotFound is returned when a document is not found
var ErrDocumentNotFound = errors.New("document not found")

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
	docRepo    *repositories.DocumentRepository
	storageRepo repositories.StorageRepository
}

func NewDocumentService(docRepo *repositories.DocumentRepository, storageRepo repositories.StorageRepository) *DocumentService {
	return &DocumentService{docRepo: docRepo, storageRepo: storageRepo}
}

func (s *DocumentService) GetDocumentsByEntity(entityType string, entityID int64, tenantID *int64, page, pageSize int, orderBy, order string) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	result, err := s.docRepo.FindByEntity(entityType, entityID, tenantID, params)
	if err != nil {
		return nil, err
	}

	items := toDocumentResponses(result.Items)
	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

func toDocumentResponses(dtos []repositories.DocumentDTO) []schemas.DocumentResponse {
	result := make([]schemas.DocumentResponse, len(dtos))
	for i := range dtos {
		result[i] = *toDocumentResponse(&dtos[i])
	}
	return result
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
func (s *DocumentService) GetDocumentByID(id int64) (*schemas.DocumentResponse, error) {
	dto, err := s.docRepo.FindDocumentByID(id)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil
	}
	return toDocumentResponse(dto), nil
}

func toDocumentResponse(dto *repositories.DocumentDTO) *schemas.DocumentResponse {
	return &schemas.DocumentResponse{
		ID:          dto.ID,
		TenantID:    dto.TenantID,
		Filename:    dto.Filename,
		ContentType: dto.ContentType,
		Description: dto.Description,
		FileSize:    dto.FileSize,
	}
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
	if err := s.storageRepo.Upload(storagePath, input.Data, input.ContentType, input.Size); err != nil {
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
		s.storageRepo.Delete(storagePath)
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

// GetDocumentVersions returns all versions of a document
func (s *DocumentService) GetDocumentVersions(documentID int64, tenantID int64) ([]repositories.DocumentDTO, int64, error) {
	versions, err := s.docRepo.FindDocumentVersions(documentID, tenantID)
	if err != nil {
		return nil, 0, err
	}
	count := int64(len(versions))
	return versions, count, nil
}

// CreateDocumentVersionInput holds data for creating a new version
type CreateDocumentVersionInput struct {
	DocumentID  int64
	TenantID    int64
	UserID      int64
	Filename    string
	ContentType string
	Data        io.Reader
	Size        int64
}

// CreateDocumentVersion creates a new version of an existing document
func (s *DocumentService) CreateDocumentVersion(input CreateDocumentVersionInput) (*repositories.DocumentDTO, int64, error) {
	if input.Size > MaxFileSize {
		return nil, 0, fmt.Errorf("file exceeds 10MB limit")
	}

	// Verify the parent document exists
	parent, err := s.docRepo.FindDocumentByID(input.DocumentID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find parent document: %w", err)
	}
	if parent == nil {
		return nil, 0, ErrDocumentNotFound
	}
	if parent.TenantID != input.TenantID {
		return nil, 0, ErrDocumentNotFound
	}

	// Generate storage path
	timestamp := time.Now().UnixMilli()
	storagePath := fmt.Sprintf("%d/%d/%s", input.TenantID, timestamp, input.Filename)

	// Upload to storage
	if err := s.storageRepo.Upload(storagePath, input.Data, input.ContentType, input.Size); err != nil {
		return nil, 0, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create new version in database
	doc, err := s.docRepo.CreateNewVersion(repositories.CreateVersionInput{
		ParentID:    input.DocumentID,
		TenantID:    input.TenantID,
		UserID:      input.UserID,
		StoragePath: storagePath,
		FileSize:    input.Size,
		ContentType: input.ContentType,
	})
	if err != nil {
		s.storageRepo.Delete(storagePath)
		return nil, 0, fmt.Errorf("failed to create version: %w", err)
	}

	// Get version count
	count, _ := s.docRepo.GetVersionCount(doc.ID, input.TenantID)

	return doc, count, nil
}

// SetCurrentVersion marks a specific version as the current version
func (s *DocumentService) SetCurrentVersion(documentID int64, tenantID int64) (*repositories.DocumentDTO, int64, error) {
	doc, err := s.docRepo.SetCurrentVersion(documentID, tenantID)
	if err != nil {
		return nil, 0, err
	}
	if doc == nil {
		return nil, 0, ErrDocumentNotFound
	}

	count, _ := s.docRepo.GetVersionCount(documentID, tenantID)

	return doc, count, nil
}

// DownloadDocumentResult holds the result of a download request
type DownloadDocumentResult struct {
	Document    *repositories.DocumentDTO
	RedirectURL *string
	Content     []byte
}

// DownloadDocument gets a document's content for download
func (s *DocumentService) DownloadDocument(documentID int64, tenantID int64) (*DownloadDocumentResult, error) {
	doc, err := s.docRepo.FindDocumentByID(documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil || doc.TenantID != tenantID {
		return nil, ErrDocumentNotFound
	}

	// Get URL from storage backend - R2 returns presigned URL, local returns file URL
	url := s.storageRepo.GetURL(doc.StoragePath)
	if url != "" && !strings.HasPrefix(url, "file://") {
		// It's a remote URL (presigned), use redirect
		return &DownloadDocumentResult{
			Document:    doc,
			RedirectURL: &url,
		}, nil
	}

	// Fall back to downloading content (local storage)
	content, err := s.storageRepo.Download(doc.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return &DownloadDocumentResult{
		Document: doc,
		Content:  content,
	}, nil
}

// UpdateDocumentInput holds data for updating a document
type UpdateDocumentInput struct {
	Description *string `json:"description"`
}

// UpdateDocument updates a document's metadata
func (s *DocumentService) UpdateDocument(documentID int64, tenantID int64, userID int64, input UpdateDocumentInput) (*repositories.DocumentDTO, error) {
	doc, err := s.docRepo.FindDocumentByID(documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil || doc.TenantID != tenantID {
		return nil, ErrDocumentNotFound
	}

	updates := map[string]interface{}{
		"updated_by": userID,
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	if err := s.docRepo.UpdateDocument(documentID, updates); err != nil {
		return nil, err
	}

	return s.docRepo.FindDocumentByID(documentID)
}

// DeleteDocument deletes a document and its storage file
func (s *DocumentService) DeleteDocument(documentID int64, tenantID int64) error {
	doc, err := s.docRepo.FindDocumentByID(documentID)
	if err != nil {
		return err
	}
	if doc == nil || doc.TenantID != tenantID {
		return ErrDocumentNotFound
	}

	// Delete from storage
	if doc.StoragePath != "" {
		s.storageRepo.Delete(doc.StoragePath)
	}

	// Delete from database
	return s.docRepo.DeleteDocument(documentID)
}
