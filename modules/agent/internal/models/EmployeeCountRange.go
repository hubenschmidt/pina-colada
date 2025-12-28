package models

import "time"

type EmployeeCountRange struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Label        string    `gorm:"uniqueIndex;not null" json:"label"`
	MinValue     *int      `json:"min_value"`
	MaxValue     *int      `json:"max_value"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (EmployeeCountRange) TableName() string {
	return "Employee_Count_Range"
}
