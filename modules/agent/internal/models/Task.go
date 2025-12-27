package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Task struct {
	ID                     int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID               *int64           `gorm:"index" json:"tenant_id"`
	TaskableType           *string          `gorm:"size:50" json:"taskable_type"` // Deal, Lead, Job, Project, Organization, Individual, Contact
	TaskableID             *int64           `json:"taskable_id"`
	Title                  string           `gorm:"not null" json:"title"`
	Description            *string          `json:"description"`
	CurrentStatusID        *int64           `gorm:"index" json:"current_status_id"`
	PriorityID             *int64           `gorm:"index" json:"priority_id"`
	StartDate              *time.Time       `gorm:"type:date" json:"start_date"`
	DueDate                *time.Time       `gorm:"type:date" json:"due_date"`
	EstimatedHours         *decimal.Decimal `gorm:"type:numeric(6,2)" json:"estimated_hours"`
	ActualHours            *decimal.Decimal `gorm:"type:numeric(6,2)" json:"actual_hours"`
	Complexity             *int16           `json:"complexity"` // Fibonacci: 1, 2, 3, 5, 8, 13, 21
	SortOrder              int              `gorm:"default:0" json:"sort_order"`
	CompletedAt            *time.Time       `json:"completed_at"`
	AssignedToIndividualID *int64           `gorm:"index" json:"assigned_to_individual_id"`
	CreatedAt              time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy              int64            `gorm:"not null" json:"created_by"`
	UpdatedBy              int64            `gorm:"not null" json:"updated_by"`

	// Relations
	CurrentStatus *Status     `gorm:"foreignKey:CurrentStatusID" json:"current_status,omitempty"`
	Priority      *Status     `gorm:"foreignKey:PriorityID" json:"priority,omitempty"`
	AssignedTo    *Individual `gorm:"foreignKey:AssignedToIndividualID" json:"assigned_to,omitempty"`
}

func (Task) TableName() string {
	return "Task"
}
