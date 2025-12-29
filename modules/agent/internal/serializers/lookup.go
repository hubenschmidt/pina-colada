package serializers

import (
	"time"

	"agent/internal/models"
)

// IndustryResponse represents an industry
type IndustryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IndustryToResponse converts a model to response
func IndustryToResponse(m *models.Industry) IndustryResponse {
	return IndustryResponse{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// IndustriesToResponse converts a slice of models to responses
func IndustriesToResponse(models []models.Industry) []IndustryResponse {
	resp := make([]IndustryResponse, len(models))
	for i := range models {
		resp[i] = IndustryToResponse(&models[i])
	}
	return resp
}

// EmployeeCountRangeResponse represents an employee count range
type EmployeeCountRangeResponse struct {
	ID           int64     `json:"id"`
	Label        string    `json:"label"`
	MinValue     *int      `json:"min_value"`
	MaxValue     *int      `json:"max_value"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EmployeeCountRangeToResponse converts a model to response
func EmployeeCountRangeToResponse(m *models.EmployeeCountRange) EmployeeCountRangeResponse {
	return EmployeeCountRangeResponse{
		ID:           m.ID,
		Label:        m.Label,
		MinValue:     m.MinValue,
		MaxValue:     m.MaxValue,
		DisplayOrder: m.DisplayOrder,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// EmployeeCountRangesToResponse converts a slice of models to responses
func EmployeeCountRangesToResponse(models []models.EmployeeCountRange) []EmployeeCountRangeResponse {
	resp := make([]EmployeeCountRangeResponse, len(models))
	for i := range models {
		resp[i] = EmployeeCountRangeToResponse(&models[i])
	}
	return resp
}

// RevenueRangeResponse represents a revenue range
type RevenueRangeResponse struct {
	ID           int64     `json:"id"`
	Label        string    `json:"label"`
	MinValue     *int64    `json:"min_value"`
	MaxValue     *int64    `json:"max_value"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RevenueRangeToResponse converts a model to response
func RevenueRangeToResponse(m *models.RevenueRange) RevenueRangeResponse {
	return RevenueRangeResponse{
		ID:           m.ID,
		Label:        m.Label,
		MinValue:     m.MinValue,
		MaxValue:     m.MaxValue,
		DisplayOrder: m.DisplayOrder,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// RevenueRangesToResponse converts a slice of models to responses
func RevenueRangesToResponse(models []models.RevenueRange) []RevenueRangeResponse {
	resp := make([]RevenueRangeResponse, len(models))
	for i := range models {
		resp[i] = RevenueRangeToResponse(&models[i])
	}
	return resp
}

// FundingStageResponse represents a funding stage
type FundingStageResponse struct {
	ID           int64     `json:"id"`
	Label        string    `json:"label"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FundingStageToResponse converts a model to response
func FundingStageToResponse(m *models.FundingStage) FundingStageResponse {
	return FundingStageResponse{
		ID:           m.ID,
		Label:        m.Label,
		DisplayOrder: m.DisplayOrder,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// FundingStagesToResponse converts a slice of models to responses
func FundingStagesToResponse(models []models.FundingStage) []FundingStageResponse {
	resp := make([]FundingStageResponse, len(models))
	for i := range models {
		resp[i] = FundingStageToResponse(&models[i])
	}
	return resp
}

// SalaryRangeResponse represents a salary range
type SalaryRangeResponse struct {
	ID           int64     `json:"id"`
	Label        string    `json:"label"`
	MinValue     *int      `json:"min_value"`
	MaxValue     *int      `json:"max_value"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SalaryRangeToResponse converts a model to response
func SalaryRangeToResponse(m *models.SalaryRange) SalaryRangeResponse {
	return SalaryRangeResponse{
		ID:           m.ID,
		Label:        m.Label,
		MinValue:     m.MinValue,
		MaxValue:     m.MaxValue,
		DisplayOrder: m.DisplayOrder,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// SalaryRangesToResponse converts a slice of models to responses
func SalaryRangesToResponse(models []models.SalaryRange) []SalaryRangeResponse {
	resp := make([]SalaryRangeResponse, len(models))
	for i := range models {
		resp[i] = SalaryRangeToResponse(&models[i])
	}
	return resp
}
