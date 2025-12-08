package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
)

// JobListResponse represents a job in list view
type JobListResponse struct {
	ID          int64      `json:"id"`
	JobTitle    string     `json:"job_title"`
	JobURL      *string    `json:"job_url"`
	Description *string    `json:"description"`
	ResumeDate  *time.Time `json:"resume_date"`
	SalaryRange *string    `json:"salary_range"`
	Status      *string    `json:"status"`
	Account     *AccountBrief `json:"account"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// JobDetailResponse represents a job in detail view
type JobDetailResponse struct {
	JobListResponse
	Lead     *LeadResponse    `json:"lead"`
	Projects []ProjectBrief   `json:"projects"`
}

// LeadResponse represents a lead
type LeadResponse struct {
	ID            int64        `json:"id"`
	Title         string       `json:"title"`
	Description   *string      `json:"description"`
	Source        *string      `json:"source"`
	Type          string       `json:"type"`
	CurrentStatus *StatusBrief `json:"current_status"`
	TenantID      *int64       `json:"tenant_id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// AccountBrief represents account summary
type AccountBrief struct {
	ID           int64               `json:"id"`
	Name         string              `json:"name"`
	Type         *string             `json:"type"`
	Organization *OrganizationBrief  `json:"organization,omitempty"`
	Individual   *IndividualBrief    `json:"individual,omitempty"`
}

// OrganizationBrief represents org summary
type OrganizationBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// IndividualBrief represents individual summary
type IndividualBrief struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     *string `json:"email"`
}

// StatusBrief represents status summary
type StatusBrief struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Category *string `json:"category"`
}

// ProjectBrief represents project summary
type ProjectBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// JobToListResponse converts a Job model to list response
func JobToListResponse(job *models.Job) JobListResponse {
	resp := JobListResponse{
		ID:          job.ID,
		JobTitle:    job.JobTitle,
		JobURL:      job.JobURL,
		Description: job.Description,
		ResumeDate:  job.ResumeDate,
		SalaryRange: job.SalaryRange,
	}

	if job.Lead != nil {
		resp.CreatedAt = job.Lead.CreatedAt
		resp.UpdatedAt = job.Lead.UpdatedAt

		if job.Lead.CurrentStatus != nil {
			resp.Status = &job.Lead.CurrentStatus.Name
		}

		if job.Lead.Account != nil {
			resp.Account = accountToResponse(job.Lead.Account)
		}
	}

	return resp
}

// JobToDetailResponse converts a Job model to detail response
func JobToDetailResponse(job *models.Job) JobDetailResponse {
	resp := JobDetailResponse{
		JobListResponse: JobToListResponse(job),
	}

	if job.Lead != nil {
		resp.Lead = leadToResponse(job.Lead)

		if job.Lead.Projects != nil {
			resp.Projects = make([]ProjectBrief, len(job.Lead.Projects))
			for i, p := range job.Lead.Projects {
				resp.Projects[i] = ProjectBrief{ID: p.ID, Name: p.Name}
			}
		}
	}

	return resp
}

func accountToResponse(account *models.Account) *AccountBrief {
	if account == nil {
		return nil
	}

	resp := &AccountBrief{
		ID:   account.ID,
		Name: account.Name,
	}

	// Derive type and details from org or individual
	if len(account.Organizations) > 0 {
		org := account.Organizations[0]
		orgType := "organization"
		resp.Type = &orgType
		resp.Name = org.Name
		resp.Organization = &OrganizationBrief{ID: org.ID, Name: org.Name}
	} else if len(account.Individuals) > 0 {
		ind := account.Individuals[0]
		indType := "individual"
		resp.Type = &indType
		resp.Name = ind.FirstName + " " + ind.LastName
		resp.Individual = &IndividualBrief{
			ID:        ind.ID,
			FirstName: ind.FirstName,
			LastName:  ind.LastName,
			Email:     ind.Email,
		}
	}

	return resp
}

func leadToResponse(lead *models.Lead) *LeadResponse {
	if lead == nil {
		return nil
	}

	resp := &LeadResponse{
		ID:          lead.ID,
		Title:       lead.Title,
		Description: lead.Description,
		Source:      lead.Source,
		Type:        lead.Type,
		TenantID:    lead.TenantID,
		CreatedAt:   lead.CreatedAt,
		UpdatedAt:   lead.UpdatedAt,
	}

	if lead.CurrentStatus != nil {
		resp.CurrentStatus = &StatusBrief{
			ID:       lead.CurrentStatus.ID,
			Name:     lead.CurrentStatus.Name,
			Category: lead.CurrentStatus.Category,
		}
	}

	return resp
}
