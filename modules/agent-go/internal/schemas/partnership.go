package schemas

// PartnershipCreate represents the request body for creating a partnership
type PartnershipCreate struct {
	AccountType     string    `json:"account_type" validate:"omitempty,oneof=Organization Individual"`
	Account         *string   `json:"account"`
	Contacts        []Contact `json:"contacts"`
	Industry        []string  `json:"industry"`
	IndustryIDs     []int64   `json:"industry_ids"`
	Title           string    `json:"title" validate:"required"`
	PartnershipName string    `json:"partnership_name" validate:"required"`
	PartnershipType *string   `json:"partnership_type"`
	StartDate       *string   `json:"start_date"`
	EndDate         *string   `json:"end_date"`
	Description     *string   `json:"description"`
	Status          string    `json:"status"`
	Source          string    `json:"source" validate:"omitempty,oneof=manual inbound referral event campaign agent"`
	ProjectIDs      []int64   `json:"project_ids"`
}

// PartnershipUpdate represents the request body for updating a partnership
type PartnershipUpdate struct {
	Account         *string   `json:"account"`
	Contacts        []Contact `json:"contacts"`
	Title           *string   `json:"title"`
	PartnershipName *string   `json:"partnership_name"`
	PartnershipType *string   `json:"partnership_type"`
	StartDate       *string   `json:"start_date"`
	EndDate         *string   `json:"end_date"`
	Description     *string   `json:"description"`
	Status          *string   `json:"status"`
	Source          *string   `json:"source"`
	ProjectIDs      []int64   `json:"project_ids"`
}
