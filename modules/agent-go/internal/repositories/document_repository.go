package repositories

import (
	"errors"
	"fmt"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

type DocumentDTO struct {
	ID               int64           `json:"id"`
	AssetType        string          `json:"asset_type"`
	TenantID         int64           `json:"tenant_id"`
	UserID           int64           `json:"user_id"`
	Filename         string          `json:"filename"`
	ContentType      string          `json:"content_type"`
	Description      *string         `json:"description"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
	StoragePath      string          `json:"storage_path"`
	FileSize         int64           `json:"file_size"`
	IsCurrentVersion bool            `json:"is_current_version"`
	VersionNumber    int             `json:"version_number"`
	ParentID         *int64          `json:"parent_id"`
	Entities         []EntityLinkDTO `json:"entities" gorm:"-"`
	Tags             []string        `json:"tags" gorm:"-"`
}

type EntityLinkDTO struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	EntityName string `json:"entity_name"`
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
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
		return nil, apperrors.ErrNotFound
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

	r.populateDocumentEntitiesAndTags(docs)

	return &PaginatedResult[DocumentDTO]{
		Items:      docs,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindDocumentVersions returns all versions of a document (whether id is root or child)
func (r *DocumentRepository) FindDocumentVersions(documentID int64, tenantID int64) ([]DocumentDTO, error) {
	// First get the document to determine root ID
	var asset models.Asset
	err := r.db.Where("id = ? AND tenant_id = ?", documentID, tenantID).First(&asset).Error
	if err != nil {
		return nil, err
	}

	// Determine root ID
	rootID := documentID
	if asset.ParentID != nil {
		rootID = *asset.ParentID
	}

	// Find all versions: root + all children
	var docs []DocumentDTO
	err = r.db.Table(`"Asset"`).
		Select(`"Asset".id, "Asset".asset_type, "Asset".tenant_id, "Asset".user_id,
			"Asset".filename, "Asset".content_type, "Asset".description,
			"Asset".created_at, "Asset".updated_at, "Asset".is_current_version, "Asset".version_number,
			"Asset".parent_id,
			"Document".storage_path, "Document".file_size`).
		Joins(`INNER JOIN "Document" ON "Document".id = "Asset".id`).
		Where(`"Asset".tenant_id = ? AND ("Asset".id = ? OR "Asset".parent_id = ?)`, tenantID, rootID, rootID).
		Order("version_number DESC").
		Scan(&docs).Error

	return docs, err
}

// GetVersionCount returns the total number of versions for a document
func (r *DocumentRepository) GetVersionCount(documentID int64, tenantID int64) (int64, error) {
	var asset models.Asset
	err := r.db.Where("id = ? AND tenant_id = ?", documentID, tenantID).First(&asset).Error
	if err != nil {
		return 0, err
	}

	rootID := documentID
	if asset.ParentID != nil {
		rootID = *asset.ParentID
	}

	var count int64
	err = r.db.Table(`"Asset"`).
		Where(`tenant_id = ? AND (id = ? OR parent_id = ?)`, tenantID, rootID, rootID).
		Count(&count).Error

	return count, err
}

// CreateVersionInput holds data for creating a new version
type CreateVersionInput struct {
	ParentID    int64
	TenantID    int64
	UserID      int64
	StoragePath string
	FileSize    int64
	ContentType string
}

// CreateNewVersion creates a new version of an existing document
func (r *DocumentRepository) CreateNewVersion(input CreateVersionInput) (*DocumentDTO, error) {
	tx := r.db.Begin()

	// Get the parent document to copy metadata
	var parentAsset models.Asset
	err := tx.Where("id = ? AND tenant_id = ?", input.ParentID, input.TenantID).First(&parentAsset).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Determine the actual root ID
	rootID := input.ParentID
	if parentAsset.ParentID != nil {
		rootID = *parentAsset.ParentID
	}

	// Get max version number
	var maxVersion int
	err = tx.Table(`"Asset"`).
		Select("COALESCE(MAX(version_number), 0)").
		Where(`tenant_id = ? AND (id = ? OR parent_id = ?)`, input.TenantID, rootID, rootID).
		Scan(&maxVersion).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Mark all existing versions as not current
	err = tx.Table(`"Asset"`).
		Where(`tenant_id = ? AND (id = ? OR parent_id = ?)`, input.TenantID, rootID, rootID).
		Update("is_current_version", false).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create new Asset (version)
	newAsset := &models.Asset{
		AssetType:        "document",
		TenantID:         input.TenantID,
		UserID:           input.UserID,
		Filename:         parentAsset.Filename,
		ContentType:      input.ContentType,
		Description:      parentAsset.Description,
		CreatedBy:        input.UserID,
		UpdatedBy:        input.UserID,
		ParentID:         &rootID,
		VersionNumber:    maxVersion + 1,
		IsCurrentVersion: true,
	}

	if err := tx.Create(newAsset).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create Document record
	doc := &models.Document{
		ID:          newAsset.ID,
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
		ID:               newAsset.ID,
		AssetType:        newAsset.AssetType,
		TenantID:         newAsset.TenantID,
		UserID:           newAsset.UserID,
		Filename:         newAsset.Filename,
		ContentType:      newAsset.ContentType,
		Description:      newAsset.Description,
		CreatedAt:        newAsset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        newAsset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		StoragePath:      doc.StoragePath,
		FileSize:         doc.FileSize,
		IsCurrentVersion: newAsset.IsCurrentVersion,
		VersionNumber:    newAsset.VersionNumber,
		ParentID:         newAsset.ParentID,
	}, nil
}

// UpdateDocument updates a document's metadata
func (r *DocumentRepository) UpdateDocument(documentID int64, updates map[string]interface{}) error {
	return r.db.Table(`"Asset"`).Where("id = ?", documentID).Updates(updates).Error
}

// DeleteDocument deletes a document and its associated records
func (r *DocumentRepository) DeleteDocument(documentID int64) error {
	tx := r.db.Begin()

	// Delete entity links
	if err := tx.Where("asset_id = ?", documentID).Delete(&models.EntityAsset{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete Document record
	if err := tx.Where("id = ?", documentID).Delete(&models.Document{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete Asset record
	if err := tx.Where("id = ?", documentID).Delete(&models.Asset{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// SetCurrentVersion marks a specific version as current and others as not current
func (r *DocumentRepository) SetCurrentVersion(documentID int64, tenantID int64) (*DocumentDTO, error) {
	tx := r.db.Begin()

	// Get the document
	var asset models.Asset
	err := tx.Where("id = ? AND tenant_id = ?", documentID, tenantID).First(&asset).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Determine root ID
	rootID := documentID
	if asset.ParentID != nil {
		rootID = *asset.ParentID
	}

	// Mark all versions as not current
	err = tx.Table(`"Asset"`).
		Where(`tenant_id = ? AND (id = ? OR parent_id = ?)`, tenantID, rootID, rootID).
		Update("is_current_version", false).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Mark the target version as current
	err = tx.Table(`"Asset"`).
		Where("id = ?", documentID).
		Update("is_current_version", true).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Fetch and return the updated document
	return r.FindDocumentByID(documentID)
}

// GetDocumentsEntities returns linked entities for multiple documents
func (r *DocumentRepository) GetDocumentsEntities(docIDs []int64) map[int64][]EntityLinkDTO {
	result := make(map[int64][]EntityLinkDTO)
	if len(docIDs) == 0 {
		return result
	}

	var links []struct {
		AssetID    int64  `gorm:"column:asset_id"`
		EntityType string `gorm:"column:entity_type"`
		EntityID   int64  `gorm:"column:entity_id"`
	}
	r.db.Table(`"Entity_Asset"`).
		Select("asset_id, entity_type, entity_id").
		Where("asset_id IN ?", docIDs).
		Scan(&links)

	// Group by document and collect unique entity lookups
	entityLookups := make(map[string]string) // "Type:ID" -> name
	for _, link := range links {
		key := fmt.Sprintf("%s:%d", link.EntityType, link.EntityID)
		entityLookups[key] = ""
	}

	// Batch fetch entity names
	r.batchFetchEntityNames(links, entityLookups)

	// Build result map
	for _, link := range links {
		key := fmt.Sprintf("%s:%d", link.EntityType, link.EntityID)
		name := entityLookups[key]
		if name == "" {
			name = fmt.Sprintf("%s #%d", link.EntityType, link.EntityID)
		}
		result[link.AssetID] = append(result[link.AssetID], EntityLinkDTO{
			EntityType: link.EntityType,
			EntityID:   link.EntityID,
			EntityName: name,
		})
	}

	return result
}

func (r *DocumentRepository) batchFetchEntityNames(links []struct {
	AssetID    int64  `gorm:"column:asset_id"`
	EntityType string `gorm:"column:entity_type"`
	EntityID   int64  `gorm:"column:entity_id"`
}, entityLookups map[string]string) {
	// Collect IDs by type
	idsByType := make(map[string][]int64)
	for _, link := range links {
		idsByType[link.EntityType] = append(idsByType[link.EntityType], link.EntityID)
	}

	tableSelectMap := map[string]struct {
		table  string
		column string
	}{
		"Organization": {`"Organization"`, "id, name"},
		"Project":      {`"Project"`, "id, name"},
		"Individual":   {`"Individual"`, "id, first_name || ' ' || last_name as name"},
		"Contact":      {`"Contact"`, "id, first_name || ' ' || last_name as name"},
	}

	for entityType, ids := range idsByType {
		r.fetchEntityNames(entityType, ids, tableSelectMap, entityLookups)
	}
}

func (r *DocumentRepository) fetchEntityNames(entityType string, ids []int64, tableSelectMap map[string]struct {
	table  string
	column string
}, entityLookups map[string]string) {
	if len(ids) == 0 {
		return
	}

	if entityType == "Lead" {
		r.fetchLeadNames(ids, entityLookups)
		return
	}

	cfg, ok := tableSelectMap[entityType]
	if !ok {
		return
	}

	var rows []struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	r.db.Table(cfg.table).Select(cfg.column).Where("id IN ?", ids).Scan(&rows)

	for _, row := range rows {
		key := fmt.Sprintf("%s:%d", entityType, row.ID)
		entityLookups[key] = row.Name
	}
}

func (r *DocumentRepository) fetchLeadNames(ids []int64, entityLookups map[string]string) {
	// Try Job
	var jobs []struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:job_title"`
	}
	r.db.Table(`"Job"`).Select("id, job_title").Where("id IN ?", ids).Scan(&jobs)
	for _, j := range jobs {
		entityLookups[fmt.Sprintf("Lead:%d", j.ID)] = j.Name
	}

	// Try Opportunity
	var opps []struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:opportunity_name"`
	}
	r.db.Table(`"Opportunity"`).Select("id, opportunity_name").Where("id IN ?", ids).Scan(&opps)
	for _, o := range opps {
		entityLookups[fmt.Sprintf("Lead:%d", o.ID)] = o.Name
	}

	// Try Partnership
	var parts []struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:partnership_name"`
	}
	r.db.Table(`"Partnership"`).Select("id, partnership_name").Where("id IN ?", ids).Scan(&parts)
	for _, p := range parts {
		entityLookups[fmt.Sprintf("Lead:%d", p.ID)] = p.Name
	}
}

// GetDocumentsTags returns tags for multiple documents
func (r *DocumentRepository) GetDocumentsTags(docIDs []int64) map[int64][]string {
	result := make(map[int64][]string)
	if len(docIDs) == 0 {
		return result
	}

	var rows []struct {
		EntityID int64  `gorm:"column:entity_id"`
		TagName  string `gorm:"column:name"`
	}
	r.db.Table(`"Entity_Tag"`).
		Select(`"Entity_Tag".entity_id, "Tag".name`).
		Joins(`INNER JOIN "Tag" ON "Tag".id = "Entity_Tag".tag_id`).
		Where(`"Entity_Tag".entity_type = ? AND "Entity_Tag".entity_id IN ?`, "Asset", docIDs).
		Order(`"Entity_Tag".entity_id, "Tag".name`).
		Scan(&rows)

	for _, row := range rows {
		result[row.EntityID] = append(result[row.EntityID], row.TagName)
	}

	return result
}

func (r *DocumentRepository) populateDocumentEntitiesAndTags(docs []DocumentDTO) {
	if len(docs) == 0 {
		return
	}

	docIDs := make([]int64, len(docs))
	for i, d := range docs {
		docIDs[i] = d.ID
	}

	entitiesMap := r.GetDocumentsEntities(docIDs)
	tagsMap := r.GetDocumentsTags(docIDs)

	for i := range docs {
		docs[i].Entities = entitiesMap[docs[i].ID]
		docs[i].Tags = tagsMap[docs[i].ID]
		if docs[i].Entities == nil {
			docs[i].Entities = []EntityLinkDTO{}
		}
		if docs[i].Tags == nil {
			docs[i].Tags = []string{}
		}
	}
}
