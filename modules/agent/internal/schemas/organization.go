package schemas

// OrganizationCreate represents the request body for creating an organization
type OrganizationCreate struct {
	Name                 string   `json:"name" validate:"required"`
	Website              *string  `json:"website" validate:"omitempty,url"`
	Phone                *string  `json:"phone" validate:"omitempty,e164"`
	EmployeeCount        *int     `json:"employee_count" validate:"omitempty,min=0"`
	EmployeeCountRangeID *int64   `json:"employee_count_range_id"`
	FundingStageID       *int64   `json:"funding_stage_id"`
	Description          *string  `json:"description"`
	AccountID            *int64   `json:"account_id"`
	IndustryIDs          []int64  `json:"industry_ids"`
	ProjectIDs           []int64  `json:"project_ids"`
	RevenueRangeID       *int64   `json:"revenue_range_id"`
	FoundingYear         *int     `json:"founding_year" validate:"omitempty,min=1800,max=2100"`
	HeadquartersCity     *string  `json:"headquarters_city"`
	HeadquartersState    *string  `json:"headquarters_state"`
	HeadquartersCountry  *string  `json:"headquarters_country"`
	CompanyType          *string  `json:"company_type" validate:"omitempty,oneof=private public nonprofit government"`
	LinkedInURL          *string  `json:"linkedin_url" validate:"omitempty,url"`
	CrunchbaseURL        *string  `json:"crunchbase_url" validate:"omitempty,url"`
}

// OrganizationUpdate represents the request body for updating an organization
type OrganizationUpdate struct {
	Name                 *string  `json:"name"`
	Website              *string  `json:"website" validate:"omitempty,url"`
	Phone                *string  `json:"phone" validate:"omitempty,e164"`
	EmployeeCount        *int     `json:"employee_count" validate:"omitempty,min=0"`
	EmployeeCountRangeID *int64   `json:"employee_count_range_id"`
	FundingStageID       *int64   `json:"funding_stage_id"`
	Description          *string  `json:"description"`
	IndustryIDs          []int64  `json:"industry_ids"`
	ProjectIDs           []int64  `json:"project_ids"`
	RevenueRangeID       *int64   `json:"revenue_range_id"`
	FoundingYear         *int     `json:"founding_year" validate:"omitempty,min=1800,max=2100"`
	HeadquartersCity     *string  `json:"headquarters_city"`
	HeadquartersState    *string  `json:"headquarters_state"`
	HeadquartersCountry  *string  `json:"headquarters_country"`
	CompanyType          *string  `json:"company_type" validate:"omitempty,oneof=private public nonprofit government"`
	LinkedInURL          *string  `json:"linkedin_url" validate:"omitempty,url"`
	CrunchbaseURL        *string  `json:"crunchbase_url" validate:"omitempty,url"`
}

// OrgContactCreate represents the request body for adding a contact to an organization
type OrgContactCreate struct {
	IndividualID *int64  `json:"individual_id"`
	FirstName    string  `json:"first_name" validate:"required"`
	LastName     string  `json:"last_name" validate:"required"`
	Title        *string `json:"title"`
	Department   *string `json:"department"`
	Role         *string `json:"role"`
	Email        *string `json:"email" validate:"omitempty,email"`
	Phone        *string `json:"phone" validate:"omitempty,e164"`
	IsPrimary    bool    `json:"is_primary"`
	Notes        *string `json:"notes"`
}

// OrgContactUpdate represents the request body for updating an organization contact
type OrgContactUpdate struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Title      *string `json:"title"`
	Department *string `json:"department"`
	Role       *string `json:"role"`
	Email      *string `json:"email" validate:"omitempty,email"`
	Phone      *string `json:"phone" validate:"omitempty,e164"`
	IsPrimary  *bool   `json:"is_primary"`
	Notes      *string `json:"notes"`
}

// OrgTechnologyCreate represents the request body for adding technology to an organization
type OrgTechnologyCreate struct {
	TechnologyID int64    `json:"technology_id" validate:"required"`
	Source       *string  `json:"source"`
	Confidence   *float64 `json:"confidence" validate:"omitempty,min=0,max=1"`
}

// FundingRoundCreate represents the request body for creating a funding round
type FundingRoundCreate struct {
	RoundType     string  `json:"round_type" validate:"required"`
	Amount        *int64  `json:"amount" validate:"omitempty,min=0"`
	AnnouncedDate *string `json:"announced_date"`
	LeadInvestor  *string `json:"lead_investor"`
	SourceURL     *string `json:"source_url" validate:"omitempty,url"`
}

// SignalCreate represents the request body for creating a company signal
type SignalCreate struct {
	SignalType     string   `json:"signal_type" validate:"required"`
	Headline       string   `json:"headline" validate:"required"`
	Description    *string  `json:"description"`
	SignalDate     *string  `json:"signal_date"`
	Source         *string  `json:"source"`
	SourceURL      *string  `json:"source_url" validate:"omitempty,url"`
	Sentiment      *string  `json:"sentiment" validate:"omitempty,oneof=positive negative neutral"`
	RelevanceScore *float64 `json:"relevance_score" validate:"omitempty,min=0,max=1"`
}

// OrgRelationshipCreate represents the request body for creating an organization relationship
type OrgRelationshipCreate struct {
	ToOrganizationID int64   `json:"to_organization_id" validate:"required"`
	RelationshipType *string `json:"relationship_type"`
	Notes            *string `json:"notes"`
}
