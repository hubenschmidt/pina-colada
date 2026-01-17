package schemas

// JobCreate represents the request body for creating a job
type JobCreate struct {
	AccountType   string    `json:"account_type" validate:"omitempty,oneof=Organization Individual"`
	Account       *string   `json:"account"`
	ContactName   *string   `json:"contact_name"`
	Contacts      []Contact `json:"contacts"`
	Industry      []string  `json:"industry"`
	IndustryIDs   []int64   `json:"industry_ids"`
	JobTitle      string    `json:"job_title" validate:"required"`
	Date          *string   `json:"date"`
	JobURL        *string   `json:"job_url" validate:"omitempty,url"`
	DatePosted    *string   `json:"date_posted"`
	SalaryRange   *string   `json:"salary_range"`
	SalaryRangeID *int64    `json:"salary_range_id"`
	Description   *string   `json:"description"`
	Resume        *string   `json:"resume"`
	Status        string    `json:"status"`
	Source        string    `json:"source" validate:"omitempty,oneof=manual inbound referral event campaign agent"`
	ProjectIDs    []int64   `json:"project_ids"`
}

// JobUpdate represents the request body for updating a job
type JobUpdate struct {
	Account       *string   `json:"account"`
	Contacts      []Contact `json:"contacts"`
	JobTitle      *string   `json:"job_title"`
	Date          *string   `json:"date"`
	JobURL        *string   `json:"job_url" validate:"omitempty,url"`
	DatePosted    *string   `json:"date_posted"`
	SalaryRange   *string   `json:"salary_range"`
	SalaryRangeID *int64    `json:"salary_range_id"`
	Description   *string   `json:"description"`
	Resume        *string   `json:"resume"`
	Status        *string   `json:"status"`
	Source        *string   `json:"source"`
	LeadStatusID  *string   `json:"lead_status_id"`
	ProjectIDs    []int64   `json:"project_ids"`
}

// Contact represents a contact in job creation
type Contact struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Title     *string `json:"title"`
}
