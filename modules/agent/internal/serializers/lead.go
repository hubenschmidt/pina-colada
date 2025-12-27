package serializers

import (
	"time"

	"github.com/shopspring/decimal"
)

// OpportunityResponse represents an opportunity response
type OpportunityResponse struct {
	ID                         int64            `json:"id"`
	Account                    string           `json:"account"`
	OpportunityName            string           `json:"opportunity_name"`
	EstimatedValue             *decimal.Decimal `json:"estimated_value"`
	Probability                *int16           `json:"probability"`
	ExpectedCloseDate          *string          `json:"expected_close_date"`
	FormattedExpectedCloseDate string           `json:"formatted_expected_close_date"`
	Status                     string           `json:"status"`
	Description                *string          `json:"description"`
	CreatedAt                  time.Time        `json:"created_at"`
	UpdatedAt                  time.Time        `json:"updated_at"`
	FormattedUpdatedAt         string           `json:"formatted_updated_at"`
}

// PartnershipResponse represents a partnership response
type PartnershipResponse struct {
	ID                 int64     `json:"id"`
	Account            string    `json:"account"`
	PartnershipName    string    `json:"partnership_name"`
	PartnershipType    *string   `json:"partnership_type"`
	Status             string    `json:"status"`
	StartDate          *string   `json:"start_date"`
	FormattedStartDate string    `json:"formatted_start_date"`
	EndDate            *string   `json:"end_date"`
	FormattedEndDate   string    `json:"formatted_end_date"`
	Description        *string   `json:"description"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	FormattedUpdatedAt string    `json:"formatted_updated_at"`
}
