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
	LinkedInURL   *string `gorm:"column:linkedin_url" json:"linkedin_url"`
	CrunchbaseURL *string `gorm:"column:crunchbase_url" json:"crunchbase_url"`

	// Relations
	Account            *Account                 `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Technologies       []OrganizationTechnology `gorm:"foreignKey:OrganizationID" json:"technologies,omitempty"`
	FundingRounds      []FundingRound           `gorm:"foreignKey:OrganizationID" json:"funding_rounds,omitempty"`
	EmployeeCountRange *EmployeeCountRange      `gorm:"foreignKey:EmployeeCountRangeID" json:"employee_count_range,omitempty"`
	FundingStage       *FundingStage            `gorm:"foreignKey:FundingStageID" json:"funding_stage,omitempty"`
}

func (Organization) TableName() string {
	return "Organization"
}
