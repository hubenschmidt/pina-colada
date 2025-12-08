package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
)

// OrganizationListResponse represents an organization in list view
type OrganizationListResponse struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Domain       *string         `json:"domain"`
	Website      *string         `json:"website"`
	Description  *string         `json:"description"`
	EmployeeCount *int           `json:"employee_count"`
	Industries   []IndustryBrief `json:"industries"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// OrganizationDetailResponse represents an organization in detail view
type OrganizationDetailResponse struct {
	OrganizationListResponse
	LinkedInURL   *string            `json:"linkedin_url"`
	TwitterURL    *string            `json:"twitter_url"`
	GithubURL     *string            `json:"github_url"`
	Phone         *string            `json:"phone"`
	Founded       *int               `json:"founded"`
	Headquarters  *string            `json:"headquarters"`
	Revenue       *string            `json:"revenue"`
	Technologies  []TechnologyBrief  `json:"technologies"`
	Contacts      []ContactBrief     `json:"contacts"`
	FundingRounds []FundingBrief     `json:"funding_rounds"`
}

// IndustryBrief represents industry summary
type IndustryBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
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
		ID:           org.ID,
		Name:         org.Name,
		Domain:       org.Domain,
		Website:      org.Website,
		Description:  org.Description,
		EmployeeCount: org.EmployeeCount,
		CreatedAt:    org.CreatedAt,
		UpdatedAt:    org.UpdatedAt,
	}

	if org.Industries != nil {
		resp.Industries = make([]IndustryBrief, len(org.Industries))
		for i, ind := range org.Industries {
			resp.Industries[i] = IndustryBrief{ID: ind.ID, Name: ind.Name}
		}
	}

	return resp
}

// OrganizationToDetailResponse converts Organization model to detail response
func OrganizationToDetailResponse(org *models.Organization) OrganizationDetailResponse {
	resp := OrganizationDetailResponse{
		OrganizationListResponse: OrganizationToListResponse(org),
		LinkedInURL:             org.LinkedInURL,
		TwitterURL:              org.TwitterURL,
		GithubURL:               org.GithubURL,
		Phone:                   org.Phone,
		Founded:                 org.Founded,
		Headquarters:            org.Headquarters,
		Revenue:                 org.Revenue,
	}

	// Technologies relationship needs Technology model loaded via preload
	// Skipping for now as OrganizationTechnology is a junction table

	if org.Contacts != nil {
		resp.Contacts = make([]ContactBrief, len(org.Contacts))
		for i, c := range org.Contacts {
			resp.Contacts[i] = ContactBrief{
				ID:        c.ID,
				FirstName: c.FirstName,
				LastName:  c.LastName,
				Email:     c.Email,
				Phone:     c.Phone,
				Title:     c.Title,
				IsPrimary: c.IsPrimary,
			}
		}
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
