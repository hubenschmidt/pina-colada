package schemas

// OpportunityCreate represents the request body for creating an opportunity
type OpportunityCreate struct {
	AccountType       string    `json:"account_type" validate:"omitempty,oneof=Organization Individual"`
	Account           *string   `json:"account"`
	Contacts          []Contact `json:"contacts"`
	Industry          []string  `json:"industry"`
	IndustryIDs       []int64   `json:"industry_ids"`
	Title             string    `json:"title" validate:"required"`
	OpportunityName   string    `json:"opportunity_name" validate:"required"`
	EstimatedValue    *float64  `json:"estimated_value" validate:"omitempty,min=0"`
	Probability       *int      `json:"probability" validate:"omitempty,min=0,max=100"`
	ExpectedCloseDate *string   `json:"expected_close_date"`
	Description       *string   `json:"description"`
	Status            string    `json:"status"`
	Source            string    `json:"source" validate:"omitempty,oneof=manual inbound referral event campaign agent"`
	ProjectIDs        []int64   `json:"project_ids"`
}

// OpportunityUpdate represents the request body for updating an opportunity
type OpportunityUpdate struct {
	Account           *string   `json:"account"`
	Contacts          []Contact `json:"contacts"`
	Title             *string   `json:"title"`
	OpportunityName   *string   `json:"opportunity_name"`
	EstimatedValue    *float64  `json:"estimated_value" validate:"omitempty,min=0"`
	Probability       *int      `json:"probability" validate:"omitempty,min=0,max=100"`
	ExpectedCloseDate *string   `json:"expected_close_date"`
	Description       *string   `json:"description"`
	Status            *string   `json:"status"`
	Source            *string   `json:"source"`
	ProjectIDs        []int64   `json:"project_ids"`
}
