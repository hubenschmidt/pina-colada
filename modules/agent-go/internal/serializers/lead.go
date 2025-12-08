package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/shopspring/decimal"
)

func formatLeadDisplayDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("01/02/2006")
}

func formatLeadDisplayDatePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("01/02/2006")
}

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
	Notes                      *string          `json:"notes"`
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
	Notes              *string   `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	FormattedUpdatedAt string    `json:"formatted_updated_at"`
}

// OpportunityToResponse converts OpportunityWithLead to response
func OpportunityToResponse(opp *repositories.OpportunityWithLead) OpportunityResponse {
	resp := OpportunityResponse{
		ID:              opp.ID,
		Account:         opp.AccountName,
		OpportunityName: opp.OpportunityName,
		EstimatedValue:  opp.EstimatedValue,
		Probability:        opp.Probability,
		Status:             opp.StatusName,
		Notes:              opp.Notes,
		CreatedAt:          opp.CreatedAt,
		UpdatedAt:          opp.UpdatedAt,
		FormattedUpdatedAt: formatLeadDisplayDate(opp.UpdatedAt),
	}

	if opp.ExpectedCloseDate != nil {
		s := opp.ExpectedCloseDate.Format("2006-01-02")
		resp.ExpectedCloseDate = &s
		resp.FormattedExpectedCloseDate = formatLeadDisplayDatePtr(opp.ExpectedCloseDate)
	}

	return resp
}

// PartnershipToResponse converts PartnershipWithLead to response
func PartnershipToResponse(p *repositories.PartnershipWithLead) PartnershipResponse {
	resp := PartnershipResponse{
		ID:              p.ID,
		Account:         p.AccountName,
		PartnershipName: p.PartnershipName,
		PartnershipType: p.PartnershipType,
		Status:             p.StatusName,
		Notes:              p.Notes,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
		FormattedUpdatedAt: formatLeadDisplayDate(p.UpdatedAt),
	}

	if p.StartDate != nil {
		s := p.StartDate.Format("2006-01-02")
		resp.StartDate = &s
		resp.FormattedStartDate = formatLeadDisplayDatePtr(p.StartDate)
	}

	if p.EndDate != nil {
		s := p.EndDate.Format("2006-01-02")
		resp.EndDate = &s
		resp.FormattedEndDate = formatLeadDisplayDatePtr(p.EndDate)
	}

	return resp
}
