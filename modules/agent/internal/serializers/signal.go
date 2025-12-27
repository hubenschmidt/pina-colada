package serializers

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/shopspring/decimal"
)

// SignalResponse represents a signal in API responses
type SignalResponse struct {
	ID             int64    `json:"id"`
	SignalType     string   `json:"signal_type"`
	Headline       string   `json:"headline"`
	Description    *string  `json:"description,omitempty"`
	SignalDate     *string  `json:"signal_date,omitempty"`
	Source         *string  `json:"source,omitempty"`
	SourceURL      *string  `json:"source_url,omitempty"`
	Sentiment      *string  `json:"sentiment,omitempty"`
	RelevanceScore *float64 `json:"relevance_score,omitempty"`
}

// SignalToResponse converts a signal model to response
func SignalToResponse(s *models.Signal) SignalResponse {
	resp := SignalResponse{
		ID:          s.ID,
		SignalType:  s.SignalType,
		Headline:    s.Headline,
		Description: s.Description,
		Source:      s.Source,
		SourceURL:   s.SourceURL,
		Sentiment:   s.Sentiment,
	}

	if s.SignalDate != nil {
		dateStr := s.SignalDate.Format("2006-01-02")
		resp.SignalDate = &dateStr
	}

	if s.RelevanceScore != nil {
		f, _ := s.RelevanceScore.Float64()
		resp.RelevanceScore = &f
	}

	return resp
}

// SignalsToResponse converts a slice of signal models to responses
func SignalsToResponse(signals []models.Signal) []SignalResponse {
	result := make([]SignalResponse, len(signals))
	for i := range signals {
		result[i] = SignalToResponse(&signals[i])
	}
	return result
}

// SignalCreateInput represents input for creating a signal
type SignalCreateInput struct {
	SignalType     string
	Headline       string
	Description    *string
	SignalDate     *string
	Source         *string
	SourceURL      *string
	Sentiment      *string
	RelevanceScore *decimal.Decimal
}
