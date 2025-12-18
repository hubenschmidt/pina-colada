package repositories

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/pina-colada-co/agent-go/internal/models"
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
	ID             int64
	UserID         int64
	TenantID       int64
	Name           *string
	StartedAt      time.Time
	EndedAt        *time.Time
	ConfigSnapshot json.RawMessage
	CreatedAt      time.Time
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

// ListSessions returns recent sessions for a user
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
		ID:             s.ID,
		UserID:         s.UserID,
		TenantID:       s.TenantID,
		Name:           s.Name,
		StartedAt:      s.StartedAt,
		EndedAt:        s.EndedAt,
		ConfigSnapshot: json.RawMessage(s.ConfigSnapshot),
		CreatedAt:      s.CreatedAt,
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
