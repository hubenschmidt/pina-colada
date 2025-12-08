package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

type DocumentDTO struct {
	ID               int64   `json:"id"`
	AssetType        string  `json:"asset_type"`
	TenantID         int64   `json:"tenant_id"`
	UserID           int64   `json:"user_id"`
	Filename         string  `json:"filename"`
	ContentType      string  `json:"content_type"`
	Description      *string `json:"description"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	StoragePath      string  `json:"storage_path"`
	FileSize         int64   `json:"file_size"`
	IsCurrentVersion bool    `json:"is_current_version"`
	VersionNumber    int     `json:"version_number"`
}

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) FindByEntity(entityType string, entityID int64, tenantID *int64, params PaginationParams) (*PaginatedResult[DocumentDTO], error) {
	var docs []DocumentDTO
	var totalCount int64

	query := r.db.Table("\"Asset\"").
		Select(`"Asset".id, "Asset".asset_type, "Asset".tenant_id, "Asset".user_id,
			"Asset".filename, "Asset".content_type, "Asset".description,
			"Asset".created_at, "Asset".updated_at, "Asset".is_current_version, "Asset".version_number,
			"Document".storage_path, "Document".file_size`).
		Joins(`INNER JOIN "Document" ON "Document".id = "Asset".id`).
		Joins(`INNER JOIN "Entity_Asset" ON "Entity_Asset".asset_id = "Asset".id`).
		Where(`"Entity_Asset".entity_type = ? AND "Entity_Asset".entity_id = ?`, entityType, entityID).
		Where(`"Asset".asset_type = ?`, "document")

	if tenantID != nil {
		query = query.Where(`"Asset".tenant_id = ?`, *tenantID)
	}

	countQuery := r.db.Table("\"Asset\"").
		Joins(`INNER JOIN "Document" ON "Document".id = "Asset".id`).
		Joins(`INNER JOIN "Entity_Asset" ON "Entity_Asset".asset_id = "Asset".id`).
		Where(`"Entity_Asset".entity_type = ? AND "Entity_Asset".entity_id = ?`, entityType, entityID).
		Where(`"Asset".asset_type = ?`, "document")

	if tenantID != nil {
		countQuery = countQuery.Where(`"Asset".tenant_id = ?`, *tenantID)
	}

	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Scan(&docs).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[DocumentDTO]{
		Items:      docs,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

func (r *DocumentRepository) FindByID(id int64) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.First(&asset, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &asset, err
}

func (r *DocumentRepository) FindDocumentByID(id int64) (*DocumentDTO, error) {
	var doc DocumentDTO
	err := r.db.Table(`"Asset"`).
		Select(`"Asset".id, "Asset".asset_type, "Asset".tenant_id, "Asset".user_id,
			"Asset".filename, "Asset".content_type, "Asset".description,
			"Asset".created_at, "Asset".updated_at, "Asset".is_current_version, "Asset".version_number,
			"Document".storage_path, "Document".file_size`).
		Joins(`INNER JOIN "Document" ON "Document".id = "Asset".id`).
		Where(`"Asset".id = ? AND "Asset".asset_type = ?`, id, "document").
		Scan(&doc).Error
	if err != nil {
		return nil, err
	}
	if doc.ID == 0 {
		return nil, nil
	}
	return &doc, nil
}

// LinkToEntity links a document to an entity
func (r *DocumentRepository) LinkToEntity(assetID int64, entityType string, entityID int64) error {
	link := &models.EntityAsset{
		AssetID:    assetID,
		EntityType: entityType,
		EntityID:   entityID,
	}
	return r.db.Create(link).Error
}

// UnlinkFromEntity removes a document's link to an entity
func (r *DocumentRepository) UnlinkFromEntity(assetID int64, entityType string, entityID int64) error {
	return r.db.Where("asset_id = ? AND entity_type = ? AND entity_id = ?", assetID, entityType, entityID).
		Delete(&models.EntityAsset{}).Error
}

// CreateDocumentInput holds data for creating a new document
type CreateDocumentInput struct {
	TenantID    int64
	UserID      int64
	Filename    string
	ContentType string
	StoragePath string
	FileSize    int64
	Description *string
}

// CreateDocument creates a new Asset and Document record
func (r *DocumentRepository) CreateDocument(input CreateDocumentInput) (*DocumentDTO, error) {
	tx := r.db.Begin()

	asset := &models.Asset{
		AssetType:        "document",
		TenantID:         input.TenantID,
		UserID:           input.UserID,
		Filename:         input.Filename,
		ContentType:      input.ContentType,
		Description:      input.Description,
		CreatedBy:        input.UserID,
		UpdatedBy:        input.UserID,
		VersionNumber:    1,
		IsCurrentVersion: true,
	}

	if err := tx.Create(asset).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	doc := &models.Document{
		ID:          asset.ID,
		StoragePath: input.StoragePath,
		FileSize:    input.FileSize,
	}

	if err := tx.Create(doc).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &DocumentDTO{
		ID:               asset.ID,
		AssetType:        asset.AssetType,
		TenantID:         asset.TenantID,
		UserID:           asset.UserID,
		Filename:         asset.Filename,
		ContentType:      asset.ContentType,
		Description:      asset.Description,
		CreatedAt:        asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		StoragePath:      doc.StoragePath,
		FileSize:         doc.FileSize,
		IsCurrentVersion: asset.IsCurrentVersion,
		VersionNumber:    asset.VersionNumber,
	}, nil
}

func (r *DocumentRepository) FindAll(entityType *string, entityID *int64, tenantID *int64, params PaginationParams) (*PaginatedResult[DocumentDTO], error) {
	var docs []DocumentDTO
	var totalCount int64

	query := r.db.Table("\"Asset\"").
		Select(`"Asset".id, "Asset".asset_type, "Asset".tenant_id, "Asset".user_id,
			"Asset".filename, "Asset".content_type, "Asset".description,
			"Asset".created_at, "Asset".updated_at, "Asset".is_current_version, "Asset".version_number,
			"Document".storage_path, "Document".file_size`).
		Joins(`INNER JOIN "Document" ON "Document".id = "Asset".id`).
		Where(`"Asset".asset_type = ?`, "document")

	countQuery := r.db.Table("\"Asset\"").
		Joins(`INNER JOIN "Document" ON "Document".id = "Asset".id`).
		Where(`"Asset".asset_type = ?`, "document")

	// Apply entity filter if provided
	if entityType != nil && entityID != nil {
		query = query.Joins(`INNER JOIN "Entity_Asset" ON "Entity_Asset".asset_id = "Asset".id`).
			Where(`"Entity_Asset".entity_type = ? AND "Entity_Asset".entity_id = ?`, *entityType, *entityID)
		countQuery = countQuery.Joins(`INNER JOIN "Entity_Asset" ON "Entity_Asset".asset_id = "Asset".id`).
			Where(`"Entity_Asset".entity_type = ? AND "Entity_Asset".entity_id = ?`, *entityType, *entityID)
	}

	if tenantID != nil {
		query = query.Where(`"Asset".tenant_id = ?`, *tenantID)
		countQuery = countQuery.Where(`"Asset".tenant_id = ?`, *tenantID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(`LOWER("Asset".filename) LIKE LOWER(?)`, searchTerm)
		countQuery = countQuery.Where(`LOWER("Asset".filename) LIKE LOWER(?)`, searchTerm)
	}

	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Scan(&docs).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[DocumentDTO]{
		Items:      docs,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}
