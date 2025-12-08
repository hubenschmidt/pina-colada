package models

import "time"

type Organization struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID            *int64    `gorm:"index" json:"account_id"`
	Name                 string    `gorm:"not null" json:"name"`
	Website              *string   `json:"website"`
	Phone                *string   `json:"phone"`
	EmployeeCount        *int      `json:"employee_count"`
	EmployeeCountRangeID *int64    `json:"employee_count_range_id"`
	FundingStageID       *int64    `json:"funding_stage_id"`
	Description          *string   `json:"description"`
	Notes                *string   `json:"notes"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy            int64     `gorm:"not null" json:"created_by"`
	UpdatedBy            int64     `gorm:"not null" json:"updated_by"`

	// Firmographic fields
	RevenueRangeID      *int64  `json:"revenue_range_id"`
	FoundingYear        *int    `json:"founding_year"`
	HeadquartersCity    *string `json:"headquarters_city"`
	HeadquartersState   *string `json:"headquarters_state"`
	HeadquartersCountry *string `gorm:"default:USA" json:"headquarters_country"`
	CompanyType         *string `json:"company_type"` // private, public, nonprofit, government
	LinkedInURL         *string `json:"linkedin_url"`
	CrunchbaseURL       *string `json:"crunchbase_url"`
	Domain              *string `json:"domain"`
	TwitterURL          *string `json:"twitter_url"`
	GithubURL           *string `json:"github_url"`
	Founded             *int    `json:"founded"`
	Headquarters        *string `json:"headquarters"`
	Revenue             *string `json:"revenue"`

	// Relations
	Industries    []Industry          `gorm:"many2many:Organization_Industry;" json:"industries,omitempty"`
	Technologies  []OrganizationTechnology `gorm:"foreignKey:OrganizationID" json:"technologies,omitempty"`
	Contacts      []Contact           `gorm:"many2many:Contact_Organization;" json:"contacts,omitempty"`
	FundingRounds []FundingRound      `gorm:"foreignKey:OrganizationID" json:"funding_rounds,omitempty"`
}

func (Organization) TableName() string {
	return "Organization"
}
