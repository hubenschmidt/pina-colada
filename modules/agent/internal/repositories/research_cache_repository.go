package repositories

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/pina-colada-co/agent-go/internal/models"
)

type ResearchCacheRepository struct {
	db *gorm.DB
}

func NewResearchCacheRepository(db *gorm.DB) *ResearchCacheRepository {
	return &ResearchCacheRepository{db: db}
}

// --- DTOs ---

type ResearchCacheDTO struct {
	ID          int64
	TenantID    int64
	CacheKey    string
	CacheType   string
	QueryParams json.RawMessage
	ResultData  json.RawMessage
	ResultCount int
	CreatedBy   *int64
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type CacheUpsertInput struct {
	TenantID    int64
	CacheKey    string
	CacheType   string
	QueryParams json.RawMessage
	ResultData  json.RawMessage
	ResultCount int
	CreatedBy   *int64
	ExpiresAt   time.Time
}

// --- Repository Methods ---

func (r *ResearchCacheRepository) Lookup(tenantID int64, cacheKey string) (*ResearchCacheDTO, error) {
	var cache models.ResearchCache
	err := r.db.Where("tenant_id = ? AND cache_key = ? AND expires_at > ?", tenantID, cacheKey, time.Now()).
		First(&cache).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.toDTO(&cache), nil
}

func (r *ResearchCacheRepository) Upsert(input CacheUpsertInput) error {
	cache := models.ResearchCache{
		TenantID:    input.TenantID,
		CacheKey:    input.CacheKey,
		CacheType:   input.CacheType,
		QueryParams: datatypes.JSON(input.QueryParams),
		ResultData:  datatypes.JSON(input.ResultData),
		ResultCount: input.ResultCount,
		CreatedBy:   input.CreatedBy,
		ExpiresAt:   input.ExpiresAt,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "cache_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"result_data", "result_count", "expires_at", "query_params"}),
	}).Create(&cache).Error
}

func (r *ResearchCacheRepository) ListRecent(tenantID int64, cacheType string, limit int) ([]ResearchCacheDTO, error) {
	var caches []models.ResearchCache

	query := r.db.Where("tenant_id = ? AND expires_at > ?", tenantID, time.Now())
	if cacheType != "" {
		query = query.Where("cache_type = ?", cacheType)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&caches).Error
	if err != nil {
		return nil, err
	}

	result := make([]ResearchCacheDTO, len(caches))
	for i := range caches {
		result[i] = *r.toDTO(&caches[i])
	}
	return result, nil
}

func (r *ResearchCacheRepository) DeleteExpired() (int64, error) {
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&models.ResearchCache{})
	return result.RowsAffected, result.Error
}

// --- Helpers ---

func (r *ResearchCacheRepository) toDTO(m *models.ResearchCache) *ResearchCacheDTO {
	return &ResearchCacheDTO{
		ID:          m.ID,
		TenantID:    m.TenantID,
		CacheKey:    m.CacheKey,
		CacheType:   m.CacheType,
		QueryParams: json.RawMessage(m.QueryParams),
		ResultData:  json.RawMessage(m.ResultData),
		ResultCount: m.ResultCount,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		ExpiresAt:   m.ExpiresAt,
	}
}
