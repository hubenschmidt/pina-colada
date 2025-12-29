package repositories

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"agent/internal/models"
)

// MetricRepository handles data access for recording sessions and metrics
type MetricRepository struct {
	db *gorm.DB
}

// NewMetricRepository creates a new metric repository
func NewMetricRepository(db *gorm.DB) *MetricRepository {
	return &MetricRepository{db: db}
}

// --- Input Structs ---

// SessionCreateInput contains input for creating a recording session
type SessionCreateInput struct {
	UserID         int64
	TenantID       int64
	Name           *string
	ConfigSnapshot json.RawMessage
}

// MetricCreateInput contains input for recording a metric
type MetricCreateInput struct {
	SessionID        int64
	ConversationID   *int64
	ThreadID         *string
	StartedAt        time.Time
	EndedAt          time.Time
	DurationMs       int
	InputTokens      int
	OutputTokens     int
	TotalTokens      int
	EstimatedCostUSD *float64
	Model            string
	Provider         string
	NodeName         *string
	ConfigSnapshot   json.RawMessage
	UserMessage      *string
}

// --- DTOs ---

// RecordingSessionDTO represents a recording session for service layer
type RecordingSessionDTO struct {
	ID                int64
	UserID            int64
	TenantID          int64
	Name              *string
	StartedAt         time.Time
	EndedAt           *time.Time
	ConfigSnapshot    json.RawMessage
	MetricCount       int
	TotalTokens       int
	TotalInputTokens  int
	TotalOutputTokens int
	TotalCostUSD      float64
	TotalDurationMs   int
	CreatedAt         time.Time
}

// MetricDTO represents a metric for service layer
type MetricDTO struct {
	ID               int64
	SessionID        int64
	ConversationID   *int64
	ThreadID         *string
	StartedAt        time.Time
	EndedAt          time.Time
	DurationMs       int
	InputTokens      int
	OutputTokens     int
	TotalTokens      int
	EstimatedCostUSD *float64
	Model            string
	Provider         string
	NodeName         *string
	ConfigSnapshot   json.RawMessage
	UserMessage      *string
	CreatedAt        time.Time
}

// --- Repository Methods ---

// CreateSession creates a new recording session
func (r *MetricRepository) CreateSession(input SessionCreateInput) (*RecordingSessionDTO, error) {
	session := models.AgentRecordingSession{
		UserID:         input.UserID,
		TenantID:       input.TenantID,
		Name:           input.Name,
		StartedAt:      time.Now(),
		ConfigSnapshot: datatypes.JSON(input.ConfigSnapshot),
	}

	err := r.db.Create(&session).Error
	if err != nil {
		return nil, err
	}

	return r.sessionToDTO(&session), nil
}

// EndSession marks a session as ended
func (r *MetricRepository) EndSession(sessionID int64) error {
	now := time.Now()
	return r.db.Model(&models.AgentRecordingSession{}).
		Where("id = ?", sessionID).
		Update("ended_at", now).Error
}

// EndAllActiveSessions stops all active recording sessions (used on app startup)
func (r *MetricRepository) EndAllActiveSessions() (int64, error) {
	now := time.Now()
	result := r.db.Model(&models.AgentRecordingSession{}).
		Where("ended_at IS NULL").
		Update("ended_at", now)
	return result.RowsAffected, result.Error
}

// GetActiveSession returns the active (non-ended) session for a user
func (r *MetricRepository) GetActiveSession(userID int64) (*RecordingSessionDTO, error) {
	var session models.AgentRecordingSession
	err := r.db.Where("user_id = ? AND ended_at IS NULL", userID).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.sessionToDTO(&session), nil
}

// GetSession returns a session by ID
func (r *MetricRepository) GetSession(sessionID int64) (*RecordingSessionDTO, error) {
	var session models.AgentRecordingSession
	err := r.db.First(&session, sessionID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.sessionToDTO(&session), nil
}

// SessionPaginationParams contains pagination parameters for sessions
type SessionPaginationParams struct {
	Page          int
	PageSize      int
	SortBy        string
	SortDirection string
}

// SessionPaginatedResult contains paginated sessions with metadata
type SessionPaginatedResult struct {
	Items       []RecordingSessionDTO
	TotalItems  int64
	CurrentPage int
	PageSize    int
	TotalPages  int
}

// ListSessions returns recent sessions for a user (legacy, no pagination)
func (r *MetricRepository) ListSessions(userID int64, limit int) ([]RecordingSessionDTO, error) {
	var sessions []models.AgentRecordingSession
	err := r.db.Where("user_id = ?", userID).
		Order("started_at DESC").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	result := make([]RecordingSessionDTO, len(sessions))
	for i := range sessions {
		result[i] = *r.sessionToDTO(&sessions[i])
	}
	return result, nil
}

// ListSessionsPaginated returns paginated sessions for a user
func (r *MetricRepository) ListSessionsPaginated(userID int64, params SessionPaginationParams) (*SessionPaginatedResult, error) {
	var totalItems int64
	if err := r.db.Model(&models.AgentRecordingSession{}).Where("user_id = ?", userID).Count(&totalItems).Error; err != nil {
		return nil, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	orderClause := buildSessionOrderClause(params.SortBy, params.SortDirection)

	var sessions []models.AgentRecordingSession
	err := r.db.Where("user_id = ?", userID).
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	items := make([]RecordingSessionDTO, len(sessions))
	for i := range sessions {
		items[i] = *r.sessionToDTO(&sessions[i])
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))

	return &SessionPaginatedResult{
		Items:       items,
		TotalItems:  totalItems,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

func buildSessionOrderClause(sortBy, sortDirection string) string {
	allowedColumns := map[string]bool{
		"started_at":        true,
		"ended_at":          true,
		"metric_count":      true,
		"total_tokens":      true,
		"total_cost_usd":    true,
		"total_duration_ms": true,
	}

	column := "started_at"
	if allowedColumns[sortBy] {
		column = sortBy
	}

	direction := "DESC"
	if sortDirection == "ASC" {
		direction = "ASC"
	}

	return column + " " + direction
}

// RecordMetric creates a new metric record
func (r *MetricRepository) RecordMetric(input MetricCreateInput) (*MetricDTO, error) {
	metric := models.AgentMetric{
		SessionID:        input.SessionID,
		ConversationID:   input.ConversationID,
		ThreadID:         input.ThreadID,
		StartedAt:        input.StartedAt,
		EndedAt:          input.EndedAt,
		DurationMs:       input.DurationMs,
		InputTokens:      input.InputTokens,
		OutputTokens:     input.OutputTokens,
		TotalTokens:      input.TotalTokens,
		EstimatedCostUSD: input.EstimatedCostUSD,
		Model:            input.Model,
		Provider:         input.Provider,
		NodeName:         input.NodeName,
		ConfigSnapshot:   datatypes.JSON(input.ConfigSnapshot),
		UserMessage:      input.UserMessage,
	}

	err := r.db.Create(&metric).Error
	if err != nil {
		return nil, err
	}

	return r.metricToDTO(&metric), nil
}

// GetSessionMetrics returns all metrics for a session
func (r *MetricRepository) GetSessionMetrics(sessionID int64) ([]MetricDTO, error) {
	var metrics []models.AgentMetric
	err := r.db.Where("session_id = ?", sessionID).
		Order("started_at ASC").
		Find(&metrics).Error
	if err != nil {
		return nil, err
	}

	result := make([]MetricDTO, len(metrics))
	for i := range metrics {
		result[i] = *r.metricToDTO(&metrics[i])
	}
	return result, nil
}

// MetricPaginationParams contains pagination and sorting parameters
type MetricPaginationParams struct {
	Page          int
	PageSize      int
	SortBy        string
	SortDirection string
}

// MetricPaginatedResult contains paginated metrics with metadata
type MetricPaginatedResult struct {
	Items       []MetricDTO
	TotalItems  int64
	CurrentPage int
	PageSize    int
	TotalPages  int
}

// GetSessionMetricsPaginated returns paginated metrics for a session
func (r *MetricRepository) GetSessionMetricsPaginated(sessionID int64, params MetricPaginationParams) (*MetricPaginatedResult, error) {
	// Count total
	var totalItems int64
	if err := r.db.Model(&models.AgentMetric{}).Where("session_id = ?", sessionID).Count(&totalItems).Error; err != nil {
		return nil, err
	}

	// Validate and set defaults
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Build order clause
	orderClause := buildOrderClause(params.SortBy, params.SortDirection)

	// Query with pagination
	var metrics []models.AgentMetric
	err := r.db.Where("session_id = ?", sessionID).
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&metrics).Error
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	items := make([]MetricDTO, len(metrics))
	for i := range metrics {
		items[i] = *r.metricToDTO(&metrics[i])
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))

	return &MetricPaginatedResult{
		Items:       items,
		TotalItems:  totalItems,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

func buildOrderClause(sortBy, sortDirection string) string {
	allowedColumns := map[string]bool{
		"started_at":         true,
		"model":              true,
		"node_name":          true,
		"input_tokens":       true,
		"output_tokens":      true,
		"duration_ms":        true,
		"estimated_cost_usd": true,
	}

	column := "started_at"
	if allowedColumns[sortBy] {
		column = sortBy
	}

	direction := "DESC"
	if sortDirection == "ASC" {
		direction = "ASC"
	}

	return column + " " + direction
}

// IncrementSessionAggregates adds metric values to session totals
func (r *MetricRepository) IncrementSessionAggregates(sessionID int64, inputTokens, outputTokens, totalTokens, durationMs int, costUSD float64) error {
	return r.db.Model(&models.AgentRecordingSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"metric_count":        gorm.Expr("metric_count + 1"),
			"total_tokens":        gorm.Expr("total_tokens + ?", totalTokens),
			"total_input_tokens":  gorm.Expr("total_input_tokens + ?", inputTokens),
			"total_output_tokens": gorm.Expr("total_output_tokens + ?", outputTokens),
			"total_cost_usd":      gorm.Expr("total_cost_usd + ?", costUSD),
			"total_duration_ms":   gorm.Expr("total_duration_ms + ?", durationMs),
		}).Error
}

// GetMetricsForSessions returns metrics for multiple sessions (for comparison)
func (r *MetricRepository) GetMetricsForSessions(sessionIDs []int64) ([]MetricDTO, error) {
	var metrics []models.AgentMetric
	err := r.db.Where("session_id IN ?", sessionIDs).
		Order("session_id, started_at ASC").
		Find(&metrics).Error
	if err != nil {
		return nil, err
	}

	result := make([]MetricDTO, len(metrics))
	for i := range metrics {
		result[i] = *r.metricToDTO(&metrics[i])
	}
	return result, nil
}

// --- Helpers ---

func (r *MetricRepository) sessionToDTO(s *models.AgentRecordingSession) *RecordingSessionDTO {
	return &RecordingSessionDTO{
		ID:                s.ID,
		UserID:            s.UserID,
		TenantID:          s.TenantID,
		Name:              s.Name,
		StartedAt:         s.StartedAt,
		EndedAt:           s.EndedAt,
		ConfigSnapshot:    json.RawMessage(s.ConfigSnapshot),
		MetricCount:       s.MetricCount,
		TotalTokens:       s.TotalTokens,
		TotalInputTokens:  s.TotalInputTokens,
		TotalOutputTokens: s.TotalOutputTokens,
		TotalCostUSD:      s.TotalCostUSD,
		TotalDurationMs:   s.TotalDurationMs,
		CreatedAt:         s.CreatedAt,
	}
}

func (r *MetricRepository) metricToDTO(m *models.AgentMetric) *MetricDTO {
	return &MetricDTO{
		ID:               m.ID,
		SessionID:        m.SessionID,
		ConversationID:   m.ConversationID,
		ThreadID:         m.ThreadID,
		StartedAt:        m.StartedAt,
		EndedAt:          m.EndedAt,
		DurationMs:       m.DurationMs,
		InputTokens:      m.InputTokens,
		OutputTokens:     m.OutputTokens,
		TotalTokens:      m.TotalTokens,
		EstimatedCostUSD: m.EstimatedCostUSD,
		Model:            m.Model,
		Provider:         m.Provider,
		NodeName:         m.NodeName,
		ConfigSnapshot:   json.RawMessage(m.ConfigSnapshot),
		UserMessage:      m.UserMessage,
		CreatedAt:        m.CreatedAt,
	}
}
