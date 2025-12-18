package services

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/pina-colada-co/agent-go/internal/repositories"
)

var ErrNoActiveSession = errors.New("no active recording session")
var ErrSessionAlreadyActive = errors.New("recording session already active")

// MetricService handles business logic for recording sessions and metrics
type MetricService struct {
	metricRepo    *repositories.MetricRepository
	configService *AgentConfigService
}

// NewMetricService creates a new metric service
func NewMetricService(metricRepo *repositories.MetricRepository, configService *AgentConfigService) *MetricService {
	return &MetricService{
		metricRepo:    metricRepo,
		configService: configService,
	}
}

// --- Response Types ---

// RecordingSessionResponse represents a session in API responses
type RecordingSessionResponse struct {
	ID                int64            `json:"id"`
	UserID            int64            `json:"user_id"`
	TenantID          int64            `json:"tenant_id"`
	Name              *string          `json:"name,omitempty"`
	StartedAt         time.Time        `json:"started_at"`
	EndedAt           *time.Time       `json:"ended_at,omitempty"`
	ConfigSnapshot    *json.RawMessage `json:"config_snapshot,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	MetricCount       int              `json:"metric_count,omitempty"`
	TotalCost         float64          `json:"total_cost,omitempty"`
	TotalTokens       int              `json:"total_tokens,omitempty"`
	TotalInputTokens  int              `json:"total_input_tokens,omitempty"`
	TotalOutputTokens int              `json:"total_output_tokens,omitempty"`
	TotalDurationMs   int              `json:"total_duration_ms,omitempty"`
	AvgLatencyMs      int              `json:"avg_latency_ms,omitempty"`
}

// MetricResponse represents a metric in API responses
type MetricResponse struct {
	ID               int64            `json:"id"`
	SessionID        int64            `json:"session_id"`
	ThreadID         *string          `json:"thread_id,omitempty"`
	StartedAt        time.Time        `json:"started_at"`
	EndedAt          time.Time        `json:"ended_at"`
	DurationMs       int              `json:"duration_ms"`
	InputTokens      int              `json:"input_tokens"`
	OutputTokens     int              `json:"output_tokens"`
	TotalTokens      int              `json:"total_tokens"`
	EstimatedCostUSD *float64         `json:"estimated_cost_usd,omitempty"`
	Model            string           `json:"model"`
	Provider         string           `json:"provider"`
	NodeName         *string          `json:"node_name,omitempty"`
	ConfigSnapshot   *json.RawMessage `json:"config_snapshot,omitempty"`
	UserMessage      *string          `json:"user_message,omitempty"`
}

// SessionWithMetricsResponse includes session and its metrics
type SessionWithMetricsResponse struct {
	Session     RecordingSessionResponse `json:"session"`
	Metrics     []MetricResponse         `json:"metrics"`
	TotalItems  int64                    `json:"total_items,omitempty"`
	CurrentPage int                      `json:"current_page,omitempty"`
	PageSize    int                      `json:"page_size,omitempty"`
	TotalPages  int                      `json:"total_pages,omitempty"`
}

// MetricQueryParams contains query parameters for metrics
type MetricQueryParams struct {
	Page          int
	PageSize      int
	SortBy        string
	SortDirection string
}

// SessionQueryParams contains query parameters for sessions list
type SessionQueryParams struct {
	Page          int
	PageSize      int
	SortBy        string
	SortDirection string
}

// SessionListResponse contains paginated sessions
type SessionListResponse struct {
	Sessions    []RecordingSessionResponse `json:"sessions"`
	TotalItems  int64                      `json:"total_items"`
	CurrentPage int                        `json:"current_page"`
	PageSize    int                        `json:"page_size"`
	TotalPages  int                        `json:"total_pages"`
}

// ComparisonDataResponse contains data for comparing sessions
type ComparisonDataResponse struct {
	Sessions []SessionWithMetricsResponse `json:"sessions"`
}

// TurnMetricInput contains data for recording a turn metric
type TurnMetricInput struct {
	SessionID      int64
	ConversationID *int64
	ThreadID       string
	StartedAt      time.Time
	EndedAt        time.Time
	DurationMs     int
	InputTokens    int32
	OutputTokens   int32
	Model          string
	Provider       string
	NodeName       string
	ConfigSnapshot json.RawMessage
	UserMessage    string
}

// --- Service Methods ---

// StopAllActiveSessions ends all sessions with NULL ended_at (used on app startup)
func (s *MetricService) StopAllActiveSessions() (int64, error) {
	return s.metricRepo.EndAllActiveSessions()
}

// StartRecording starts a new recording session
func (s *MetricService) StartRecording(userID, tenantID int64) (*RecordingSessionResponse, error) {
	existing, err := s.metricRepo.GetActiveSession(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSessionAlreadyActive
	}

	configSnapshot := s.captureConfigSnapshot(userID)

	input := repositories.SessionCreateInput{
		UserID:         userID,
		TenantID:       tenantID,
		ConfigSnapshot: configSnapshot,
	}

	session, err := s.metricRepo.CreateSession(input)
	if err != nil {
		return nil, err
	}

	return s.sessionDTOToResponse(session), nil
}

// StopRecording stops the active recording session
func (s *MetricService) StopRecording(userID int64) (*RecordingSessionResponse, error) {
	session, err := s.metricRepo.GetActiveSession(userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrNoActiveSession
	}

	err = s.metricRepo.EndSession(session.ID)
	if err != nil {
		return nil, err
	}

	updated, err := s.metricRepo.GetSession(session.ID)
	if err != nil {
		return nil, err
	}

	return s.sessionDTOToResponse(updated), nil
}

// GetActiveSession returns the active session for a user (if any)
func (s *MetricService) GetActiveSession(userID int64) (*RecordingSessionResponse, error) {
	session, err := s.metricRepo.GetActiveSession(userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	return s.sessionDTOToResponse(session), nil
}

// RecordTurn records metrics for a completed agent turn
func (s *MetricService) RecordTurn(input TurnMetricInput) error {
	cost := CalculateCost(input.Model, input.InputTokens, input.OutputTokens)

	threadID := &input.ThreadID
	nodeName := &input.NodeName
	userMessage := &input.UserMessage

	inputTokens := int(input.InputTokens)
	outputTokens := int(input.OutputTokens)
	totalTokens := inputTokens + outputTokens

	metricInput := repositories.MetricCreateInput{
		SessionID:        input.SessionID,
		ConversationID:   input.ConversationID,
		ThreadID:         threadID,
		StartedAt:        input.StartedAt,
		EndedAt:          input.EndedAt,
		DurationMs:       input.DurationMs,
		InputTokens:      inputTokens,
		OutputTokens:     outputTokens,
		TotalTokens:      totalTokens,
		EstimatedCostUSD: &cost,
		Model:            input.Model,
		Provider:         input.Provider,
		NodeName:         nodeName,
		ConfigSnapshot:   input.ConfigSnapshot,
		UserMessage:      userMessage,
	}

	_, err := s.metricRepo.RecordMetric(metricInput)
	if err != nil {
		return err
	}

	return s.metricRepo.IncrementSessionAggregates(input.SessionID, inputTokens, outputTokens, totalTokens, input.DurationMs, cost)
}

// GetSessionWithMetrics returns a session with paginated metrics
func (s *MetricService) GetSessionWithMetrics(sessionID int64, params *MetricQueryParams) (*SessionWithMetricsResponse, error) {
	session, err := s.metricRepo.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Use pagination if params provided
	if params != nil && params.PageSize > 0 {
		return s.getSessionWithPaginatedMetrics(session, sessionID, params)
	}

	// Fall back to non-paginated for backwards compatibility
	return s.getSessionWithAllMetrics(session, sessionID)
}

func (s *MetricService) getSessionWithPaginatedMetrics(session *repositories.RecordingSessionDTO, sessionID int64, params *MetricQueryParams) (*SessionWithMetricsResponse, error) {
	paginatedResult, err := s.metricRepo.GetSessionMetricsPaginated(sessionID, repositories.MetricPaginationParams{
		Page:          params.Page,
		PageSize:      params.PageSize,
		SortBy:        params.SortBy,
		SortDirection: params.SortDirection,
	})
	if err != nil {
		return nil, err
	}

	sessionResp := s.sessionDTOToResponse(session)

	metricResponses := make([]MetricResponse, len(paginatedResult.Items))
	for i, m := range paginatedResult.Items {
		metricResponses[i] = *s.metricDTOToResponse(&m)
	}

	return &SessionWithMetricsResponse{
		Session:     *sessionResp,
		Metrics:     metricResponses,
		TotalItems:  paginatedResult.TotalItems,
		CurrentPage: paginatedResult.CurrentPage,
		PageSize:    paginatedResult.PageSize,
		TotalPages:  paginatedResult.TotalPages,
	}, nil
}

func (s *MetricService) getSessionWithAllMetrics(session *repositories.RecordingSessionDTO, sessionID int64) (*SessionWithMetricsResponse, error) {
	metrics, err := s.metricRepo.GetSessionMetrics(sessionID)
	if err != nil {
		return nil, err
	}

	sessionResp := s.sessionDTOToResponse(session)

	metricResponses := make([]MetricResponse, len(metrics))
	for i, m := range metrics {
		metricResponses[i] = *s.metricDTOToResponse(&m)
	}

	return &SessionWithMetricsResponse{
		Session: *sessionResp,
		Metrics: metricResponses,
	}, nil
}

// ListSessions returns recent sessions for a user (legacy)
func (s *MetricService) ListSessions(userID int64, limit int) ([]RecordingSessionResponse, error) {
	sessions, err := s.metricRepo.ListSessions(userID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]RecordingSessionResponse, len(sessions))
	for i := range sessions {
		result[i] = *s.sessionDTOToResponse(&sessions[i])
	}
	return result, nil
}

// ListSessionsPaginated returns paginated sessions for a user
func (s *MetricService) ListSessionsPaginated(userID int64, params *SessionQueryParams) (*SessionListResponse, error) {
	repoParams := repositories.SessionPaginationParams{
		Page:          params.Page,
		PageSize:      params.PageSize,
		SortBy:        params.SortBy,
		SortDirection: params.SortDirection,
	}

	result, err := s.metricRepo.ListSessionsPaginated(userID, repoParams)
	if err != nil {
		return nil, err
	}

	sessions := make([]RecordingSessionResponse, len(result.Items))
	for i := range result.Items {
		sessions[i] = *s.sessionDTOToResponse(&result.Items[i])
	}

	return &SessionListResponse{
		Sessions:    sessions,
		TotalItems:  result.TotalItems,
		CurrentPage: result.CurrentPage,
		PageSize:    result.PageSize,
		TotalPages:  result.TotalPages,
	}, nil
}

// CompareSession returns data for comparing multiple sessions
func (s *MetricService) CompareSessions(sessionIDs []int64) (*ComparisonDataResponse, error) {
	result := make([]SessionWithMetricsResponse, 0, len(sessionIDs))

	for _, id := range sessionIDs {
		sessionWithMetrics, err := s.GetSessionWithMetrics(id, nil)
		if err != nil {
			return nil, err
		}
		if sessionWithMetrics == nil {
			continue
		}
		result = append(result, *sessionWithMetrics)
	}

	return &ComparisonDataResponse{Sessions: result}, nil
}

// --- Helpers ---

func (s *MetricService) captureConfigSnapshot(userID int64) json.RawMessage {
	if s.configService == nil {
		return nil
	}

	config, err := s.configService.GetAgentConfig(userID)
	if err != nil {
		return nil
	}

	snapshot, err := json.Marshal(config)
	if err != nil {
		return nil
	}

	return snapshot
}

func (s *MetricService) sessionDTOToResponse(dto *repositories.RecordingSessionDTO) *RecordingSessionResponse {
	var configSnapshot *json.RawMessage
	if len(dto.ConfigSnapshot) > 0 {
		configSnapshot = &dto.ConfigSnapshot
	}

	avgLatencyMs := 0
	if dto.MetricCount > 0 {
		avgLatencyMs = dto.TotalDurationMs / dto.MetricCount
	}

	return &RecordingSessionResponse{
		ID:                dto.ID,
		UserID:            dto.UserID,
		TenantID:          dto.TenantID,
		Name:              dto.Name,
		StartedAt:         dto.StartedAt,
		EndedAt:           dto.EndedAt,
		ConfigSnapshot:    configSnapshot,
		CreatedAt:         dto.CreatedAt,
		MetricCount:       dto.MetricCount,
		TotalCost:         dto.TotalCostUSD,
		TotalTokens:       dto.TotalTokens,
		TotalInputTokens:  dto.TotalInputTokens,
		TotalOutputTokens: dto.TotalOutputTokens,
		TotalDurationMs:   dto.TotalDurationMs,
		AvgLatencyMs:      avgLatencyMs,
	}
}

func (s *MetricService) metricDTOToResponse(dto *repositories.MetricDTO) *MetricResponse {
	var configSnapshot *json.RawMessage
	if len(dto.ConfigSnapshot) > 0 {
		configSnapshot = &dto.ConfigSnapshot
	}

	return &MetricResponse{
		ID:               dto.ID,
		SessionID:        dto.SessionID,
		ThreadID:         dto.ThreadID,
		StartedAt:        dto.StartedAt,
		EndedAt:          dto.EndedAt,
		DurationMs:       dto.DurationMs,
		InputTokens:      dto.InputTokens,
		OutputTokens:     dto.OutputTokens,
		TotalTokens:      dto.TotalTokens,
		EstimatedCostUSD: dto.EstimatedCostUSD,
		Model:            dto.Model,
		Provider:         dto.Provider,
		NodeName:         dto.NodeName,
		ConfigSnapshot:   configSnapshot,
		UserMessage:      dto.UserMessage,
	}
}
