package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
)

// IndividualListResponse represents an individual in list view
type IndividualListResponse struct {
	ID             int64           `json:"id"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          *string         `json:"email"`
	Phone          *string         `json:"phone"`
	Title          *string         `json:"title"`
	SeniorityLevel *string         `json:"seniority_level"`
	Department     *string         `json:"department"`
	IsDecisionMaker *bool          `json:"is_decision_maker"`
	Industries     []IndustryBrief `json:"industries"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// IndividualDetailResponse represents an individual in detail view
type IndividualDetailResponse struct {
	IndividualListResponse
	LinkedInURL *string        `json:"linkedin_url"`
	TwitterURL  *string        `json:"twitter_url"`
	GithubURL   *string        `json:"github_url"`
	Bio         *string        `json:"bio"`
	Description *string        `json:"description"`
	Contacts    []ContactBrief `json:"contacts"`
	Projects    []ProjectBrief `json:"projects"`
}

// IndividualToListResponse converts Individual model to list response
func IndividualToListResponse(ind *models.Individual) IndividualListResponse {
	resp := IndividualListResponse{
		ID:             ind.ID,
		FirstName:      ind.FirstName,
		LastName:       ind.LastName,
		Email:          ind.Email,
		Phone:          ind.Phone,
		Title:          ind.Title,
		SeniorityLevel: ind.SeniorityLevel,
		Department:     ind.Department,
		IsDecisionMaker: ind.IsDecisionMaker,
		CreatedAt:      ind.CreatedAt,
		UpdatedAt:      ind.UpdatedAt,
	}

	if ind.Industries != nil {
		resp.Industries = make([]IndustryBrief, len(ind.Industries))
		for i, industry := range ind.Industries {
			resp.Industries[i] = IndustryBrief{ID: industry.ID, Name: industry.Name}
		}
	}

	return resp
}

// IndividualToDetailResponse converts Individual model to detail response
func IndividualToDetailResponse(ind *models.Individual) IndividualDetailResponse {
	resp := IndividualDetailResponse{
		IndividualListResponse: IndividualToListResponse(ind),
		LinkedInURL:           ind.LinkedInURL,
		TwitterURL:            ind.TwitterURL,
		GithubURL:             ind.GithubURL,
		Bio:                   ind.Bio,
		Description:           ind.Description,
	}

	if ind.Contacts != nil {
		resp.Contacts = make([]ContactBrief, len(ind.Contacts))
		for i, c := range ind.Contacts {
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

	if ind.Projects != nil {
		resp.Projects = make([]ProjectBrief, len(ind.Projects))
		for i, p := range ind.Projects {
			resp.Projects[i] = ProjectBrief{ID: p.ID, Name: p.Name}
		}
	}

	return resp
}
