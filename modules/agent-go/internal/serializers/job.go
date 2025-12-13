package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
)

func formatDisplayDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("01/02/2006")
}

func formatDisplayDatePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("01/02/2006")
}

func formatISODate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatDateTimePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// JobListResponse represents a job in list view
type JobListResponse struct {
	ID                   int64      `json:"id"`
	JobTitle             string     `json:"job_title"`
	JobURL               *string    `json:"job_url"`
	Description          *string    `json:"description"`
	Resume               string     `json:"resume"`
	ResumeDate           *time.Time `json:"resume_date"`
	FormattedResumeDate  string     `json:"formatted_resume_date"`
	SalaryRange          *string    `json:"salary_range"`
	SalaryRangeID        *int64     `json:"salary_range_id"`
	Status               *string    `json:"status"`
	Account              string     `json:"account"`
	AccountType          string     `json:"account_type"`
	Date                 string     `json:"date"`
	FormattedDate        string     `json:"formatted_date"`
	CreatedAt            time.Time  `json:"created_at"`
	FormattedCreatedAt   string     `json:"formatted_created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	FormattedUpdatedAt   string     `json:"formatted_updated_at"`
	Source               string     `json:"source"`
	ProjectIDs           []int64    `json:"project_ids"`
}

// JobDetailResponse represents a job in detail view
type JobDetailResponse struct {
	JobListResponse
	Lead       *LeadResponse   `json:"lead"`
	Projects   []ProjectBrief  `json:"projects"`
	LeadStatus StatusResponse  `json:"lead_status,omitempty"`
}

// LeadResponse represents a lead
type LeadResponse struct {
	ID            int64        `json:"id"`
	Source        *string      `json:"source"`
	Type          string       `json:"type"`
	CurrentStatus *StatusBrief `json:"current_status"`
	TenantID      *int64       `json:"tenant_id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
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

// StatusResponse represents a full status
type StatusResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	IsTerminal  bool    `json:"is_terminal"`
}

// StatusToResponse converts a Status model to response
func StatusToResponse(status *models.Status) StatusResponse {
	if status == nil {
		return StatusResponse{}
	}
	return StatusResponse{
		ID:          status.ID,
		Name:        status.Name,
		Description: status.Description,
		Category:    status.Category,
		IsTerminal:  status.IsTerminal,
	}
}

// ProjectBrief represents project summary
type ProjectBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// JobToListResponse converts a Job model to list response
func JobToListResponse(job *models.Job) JobListResponse {
	resp := JobListResponse{
		ID:                  job.ID,
		JobTitle:            job.JobTitle,
		JobURL:              job.JobURL,
		Description:         job.Description,
		Resume:              formatDateTimePtr(job.ResumeDate),
		ResumeDate:          job.ResumeDate,
		FormattedResumeDate: formatDisplayDatePtr(job.ResumeDate),
		SalaryRange:         job.SalaryRange,
		SalaryRangeID:       job.SalaryRangeID,
		Account:             "",
		AccountType:         "Organization",
		Source:              "manual",
	}

	if job.Lead == nil {
		return resp
	}

	resp.Date = formatISODate(job.Lead.CreatedAt)
	resp.FormattedDate = formatDisplayDate(job.Lead.CreatedAt)
	resp.CreatedAt = job.Lead.CreatedAt
	resp.FormattedCreatedAt = formatDisplayDate(job.Lead.CreatedAt)
	resp.UpdatedAt = job.Lead.UpdatedAt
	resp.FormattedUpdatedAt = formatDisplayDate(job.Lead.UpdatedAt)

	if job.Lead.Source != nil {
		resp.Source = *job.Lead.Source
	}

	if job.Lead.CurrentStatus != nil {
		resp.Status = &job.Lead.CurrentStatus.Name
	}

	if job.Lead.Account != nil {
		resp.Account, resp.AccountType = extractCompanyFromAccount(job.Lead.Account)
	}

	if job.Lead.Projects != nil {
		resp.ProjectIDs = make([]int64, len(job.Lead.Projects))
		for i, p := range job.Lead.Projects {
			resp.ProjectIDs[i] = p.ID
		}
	}

	return resp
}

// extractCompanyFromAccount gets company name and type from account
func extractCompanyFromAccount(account *models.Account) (string, string) {
	if account == nil {
		return "", "Organization"
	}

	if len(account.Organizations) > 0 {
		return account.Organizations[0].Name, "Organization"
	}

	if len(account.Individuals) == 0 {
		return "", "Organization"
	}

	ind := account.Individuals[0]
	firstName := ind.FirstName
	lastName := ind.LastName

	if lastName != "" && firstName != "" {
		return lastName + ", " + firstName, "Individual"
	}
	if lastName != "" {
		return lastName, "Individual"
	}
	if firstName != "" {
		return firstName, "Individual"
	}

	return "", "Organization"
}

// JobToDetailResponse converts a Job model to detail response
func JobToDetailResponse(job *models.Job) JobDetailResponse {
	resp := JobDetailResponse{
		JobListResponse: JobToListResponse(job),
	}

	if job.Lead == nil {
		return resp
	}

	resp.Lead = leadToResponse(job.Lead)

	if job.Lead.Projects == nil {
		return resp
	}

	resp.Projects = make([]ProjectBrief, len(job.Lead.Projects))
	for i, p := range job.Lead.Projects {
		resp.Projects[i] = ProjectBrief{ID: p.ID, Name: p.Name}
	}

	return resp
}

func leadToResponse(lead *models.Lead) *LeadResponse {
	if lead == nil {
		return nil
	}

	resp := &LeadResponse{
		ID:        lead.ID,
		Source:    lead.Source,
		Type:      lead.Type,
		TenantID:  lead.TenantID,
		CreatedAt: lead.CreatedAt,
		UpdatedAt: lead.UpdatedAt,
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
