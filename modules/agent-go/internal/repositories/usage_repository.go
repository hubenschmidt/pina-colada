package repositories

import (
	"time"

	"gorm.io/gorm"
)

// UsageRepository handles usage analytics data access
type UsageRepository struct {
	db *gorm.DB
}

// NewUsageRepository creates a new usage repository
func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

// UsageRecord represents a usage analytics record for aggregation
type UsageRecord struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// TimeseriesRecord represents a time-grouped usage record
type TimeseriesRecord struct {
	Date         string `json:"date"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TotalTokens  int    `json:"total_tokens"`
}

// NodeUsageRecord represents usage breakdown by node
type NodeUsageRecord struct {
	NodeName          string `json:"node_name"`
	RequestCount      int    `json:"request_count"`
	ConversationCount int    `json:"conversation_count"`
	InputTokens       int    `json:"input_tokens"`
	OutputTokens      int    `json:"output_tokens"`
	TotalTokens       int    `json:"total_tokens"`
}

// ModelUsageRecord represents usage breakdown by model
type ModelUsageRecord struct {
	ModelName    string `json:"model_name"`
	RequestCount int    `json:"request_count"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TotalTokens  int    `json:"total_tokens"`
}

var periodOffsets = map[string][3]int{
	"daily":   {0, 0, -1},
	"weekly":  {0, 0, -7},
	"monthly": {0, -1, 0},
	"yearly":  {-1, 0, 0},
}

func (r *UsageRepository) getPeriodStart(period string) time.Time {
	now := time.Now()
	offset, ok := periodOffsets[period]
	if !ok {
		offset = periodOffsets["monthly"]
	}
	return now.AddDate(offset[0], offset[1], offset[2])
}

// GetUserUsage returns usage records for a user within the given period
func (r *UsageRepository) GetUserUsage(userID int64, period string) ([]UsageRecord, error) {
	periodStart := r.getPeriodStart(period)

	var records []UsageRecord
	err := r.db.Table(`"Usage_Analytics"`).
		Select("input_tokens, output_tokens, total_tokens").
		Where("user_id = ? AND created_at >= ?", userID, periodStart).
		Find(&records).Error

	return records, err
}

// GetTenantUsage returns usage records for a tenant within the given period
func (r *UsageRepository) GetTenantUsage(tenantID int64, period string) ([]UsageRecord, error) {
	periodStart := r.getPeriodStart(period)

	var records []UsageRecord
	err := r.db.Table(`"Usage_Analytics"`).
		Select("input_tokens, output_tokens, total_tokens").
		Where("tenant_id = ? AND created_at >= ?", tenantID, periodStart).
		Find(&records).Error

	return records, err
}

// GetUsageTimeseries returns usage data grouped by date for charting
func (r *UsageRepository) GetUsageTimeseries(tenantID int64, period string, userID *int64) ([]TimeseriesRecord, error) {
	periodStart := r.getPeriodStart(period)

	query := r.db.Table(`"Usage_Analytics"`).
		Select(`DATE(created_at) as date,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			SUM(total_tokens) as total_tokens`).
		Where("tenant_id = ? AND created_at >= ?", tenantID, periodStart).
		Group("DATE(created_at)").
		Order("date")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	var records []TimeseriesRecord
	err := query.Find(&records).Error
	return records, err
}

// GetUsageByNode returns usage breakdown by node for developer analytics
func (r *UsageRepository) GetUsageByNode(tenantID int64, period string) ([]NodeUsageRecord, error) {
	periodStart := r.getPeriodStart(period)

	var records []NodeUsageRecord
	err := r.db.Table(`"Usage_Analytics"`).
		Select(`COALESCE(node_name, 'unknown') as node_name,
			COUNT(*) as request_count,
			COUNT(DISTINCT conversation_id) as conversation_count,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			SUM(total_tokens) as total_tokens`).
		Where("tenant_id = ? AND created_at >= ?", tenantID, periodStart).
		Group("node_name").
		Order("total_tokens DESC").
		Find(&records).Error

	return records, err
}

// GetUsageByModel returns usage breakdown by model for developer analytics
func (r *UsageRepository) GetUsageByModel(tenantID int64, period string) ([]ModelUsageRecord, error) {
	periodStart := r.getPeriodStart(period)

	var records []ModelUsageRecord
	err := r.db.Table(`"Usage_Analytics"`).
		Select(`COALESCE(model_name, 'unknown') as model_name,
			COUNT(*) as request_count,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			SUM(total_tokens) as total_tokens`).
		Where("tenant_id = ? AND created_at >= ?", tenantID, periodStart).
		Group("model_name").
		Order("total_tokens DESC").
		Find(&records).Error

	return records, err
}
