package schemas

// IndividualCreate represents the request body for creating an individual
type IndividualCreate struct {
	FirstName       string  `json:"first_name" validate:"required"`
	LastName        string  `json:"last_name" validate:"required"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Phone           *string `json:"phone" validate:"omitempty,e164"`
	LinkedInURL     *string `json:"linkedin_url" validate:"omitempty,url"`
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	AccountID       *int64  `json:"account_id"`
	IndustryIDs     []int64 `json:"industry_ids"`
	ProjectIDs      []int64 `json:"project_ids"`
	TwitterURL      *string `json:"twitter_url" validate:"omitempty,url"`
	GithubURL       *string `json:"github_url" validate:"omitempty,url"`
	Bio             *string `json:"bio"`
	SeniorityLevel  *string `json:"seniority_level" validate:"omitempty,oneof=C-Level VP Director Manager IC"`
	Department      *string `json:"department"`
	IsDecisionMaker *bool   `json:"is_decision_maker"`
	ReportsToID     *int64  `json:"reports_to_id"`
}

// IndividualUpdate represents the request body for updating an individual
type IndividualUpdate struct {
	FirstName       *string `json:"first_name"`
	LastName        *string `json:"last_name"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Phone           *string `json:"phone" validate:"omitempty,e164"`
	LinkedInURL     *string `json:"linkedin_url" validate:"omitempty,url"`
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	IndustryIDs     []int64 `json:"industry_ids"`
	ProjectIDs      []int64 `json:"project_ids"`
	TwitterURL      *string `json:"twitter_url" validate:"omitempty,url"`
	GithubURL       *string `json:"github_url" validate:"omitempty,url"`
	Bio             *string `json:"bio"`
	SeniorityLevel  *string `json:"seniority_level" validate:"omitempty,oneof=C-Level VP Director Manager IC"`
	Department      *string `json:"department"`
	IsDecisionMaker *bool   `json:"is_decision_maker"`
	ReportsToID     *int64  `json:"reports_to_id"`
}

// IndContactCreate represents the request body for creating an individual's contact
type IndContactCreate struct {
	FirstName          *string `json:"first_name"`
	LastName           *string `json:"last_name"`
	OrganizationID     *int64  `json:"organization_id"`
	LinkedIndividualID *int64  `json:"linked_individual_id"`
	Title              *string `json:"title"`
	Department         *string `json:"department"`
	Role               *string `json:"role"`
	Email              *string `json:"email" validate:"omitempty,email"`
	Phone              *string `json:"phone" validate:"omitempty,e164"`
	IsPrimary          bool    `json:"is_primary"`
	Notes              *string `json:"notes"`
}

// IndContactUpdate represents the request body for updating an individual's contact
type IndContactUpdate struct {
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
	OrganizationID *int64  `json:"organization_id"`
	Title          *string `json:"title"`
	Department     *string `json:"department"`
	Role           *string `json:"role"`
	Email          *string `json:"email" validate:"omitempty,email"`
	Phone          *string `json:"phone" validate:"omitempty,e164"`
	IsPrimary      *bool   `json:"is_primary"`
	Notes          *string `json:"notes"`
}

// IndRelationshipCreate represents the request body for creating an individual relationship
type IndRelationshipCreate struct {
	ToIndividualID   int64   `json:"to_individual_id" validate:"required"`
	RelationshipType *string `json:"relationship_type"`
	Notes            *string `json:"notes"`
}
