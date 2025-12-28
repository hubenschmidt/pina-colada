package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
)

// OrganizationListResponse represents an organization in list view
type OrganizationListResponse struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Website            *string   `json:"website"`
	Description        *string   `json:"description"`
	EmployeeCountRange *string   `json:"employee_count_range"`
	FundingStage       *string   `json:"funding_stage"`
	Industries         []string  `json:"industries"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// OrganizationDetailResponse represents an organization in detail view
type OrganizationDetailResponse struct {
	OrganizationListResponse
	LinkedInURL   *string           `json:"linkedin_url"`
	CrunchbaseURL *string           `json:"crunchbase_url"`
	Phone         *string           `json:"phone"`
	Technologies  []TechnologyBrief `json:"technologies"`
	FundingRounds []FundingBrief    `json:"funding_rounds"`
}


// TechnologyBrief represents technology summary
type TechnologyBrief struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Category *string `json:"category"`
}

// ContactBrief represents contact summary
type ContactBrief struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Title     *string `json:"title"`
	IsPrimary bool    `json:"is_primary"`
}

// FundingBrief represents funding round summary
type FundingBrief struct {
	ID          int64      `json:"id"`
	RoundType   string     `json:"round_type"`
	Amount      *int64     `json:"amount"`
	AnnouncedAt *time.Time `json:"announced_at"`
}

// OrganizationToListResponse converts Organization model to list response
func OrganizationToListResponse(org *models.Organization) OrganizationListResponse {
	resp := OrganizationListResponse{
		ID:          org.ID,
		Name:        org.Name,
		Website:     org.Website,
		Description: org.Description,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}

	if org.EmployeeCountRange != nil {
		resp.EmployeeCountRange = &org.EmployeeCountRange.Label
	}

	if org.FundingStage != nil {
		resp.FundingStage = &org.FundingStage.Label
	}

	// Get industries from Account relation
	if org.Account != nil && org.Account.Industries != nil {
		resp.Industries = make([]string, len(org.Account.Industries))
		for i, ind := range org.Account.Industries {
			resp.Industries[i] = ind.Name
		}
	}

	return resp
}

// OrganizationToDetailResponse converts Organization model to detail response
func OrganizationToDetailResponse(org *models.Organization) OrganizationDetailResponse {
	resp := OrganizationDetailResponse{
		OrganizationListResponse: OrganizationToListResponse(org),
		LinkedInURL:              org.LinkedInURL,
		CrunchbaseURL:            org.CrunchbaseURL,
		Phone:                    org.Phone,
	}

	if org.FundingRounds != nil {
		resp.FundingRounds = make([]FundingBrief, len(org.FundingRounds))
		for i, f := range org.FundingRounds {
			resp.FundingRounds[i] = FundingBrief{
				ID:          f.ID,
				RoundType:   f.RoundType,
				Amount:      f.Amount,
				AnnouncedAt: f.AnnouncedDate,
			}
		}
	}

	return resp
}

// OrgTechnologyResponse represents a technology linked to an organization
type OrgTechnologyResponse struct {
	OrganizationID int64    `json:"organization_id"`
	TechnologyID   int64    `json:"technology_id"`
	Source         *string  `json:"source"`
	Confidence     *float64 `json:"confidence"`
	DetectedAt     string   `json:"detected_at"`
}

// FundingRoundResponse represents a funding round in API responses
type FundingRoundResponse struct {
	ID             int64   `json:"id"`
	OrganizationID int64   `json:"organization_id"`
	RoundType      string  `json:"round_type"`
	Amount         *int64  `json:"amount"`
	AnnouncedDate  *string `json:"announced_date"`
	LeadInvestor   *string `json:"lead_investor"`
	SourceURL      *string `json:"source_url"`
}
