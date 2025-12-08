package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/shopspring/decimal"
)

func decimalToString(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	s := d.String()
	return &s
}

// TaskResponse represents a task
type TaskResponse struct {
	ID             int64            `json:"id"`
	Title          string           `json:"title"`
	Description    *string          `json:"description"`
	TaskableType   *string          `json:"taskable_type"`
	TaskableID     *int64           `json:"taskable_id"`
	CurrentStatus  *StatusBrief     `json:"current_status"`
	Priority       *StatusBrief     `json:"priority"`
	StartDate      *time.Time       `json:"start_date"`
	DueDate        *time.Time       `json:"due_date"`
	CompletedAt    *time.Time       `json:"completed_at"`
	EstimatedHours *string          `json:"estimated_hours"` // decimal as string
	ActualHours    *string          `json:"actual_hours"`    // decimal as string
	Complexity     *int16           `json:"complexity"`
	SortOrder      int              `json:"sort_order"`
	AssignedTo     *IndividualBrief `json:"assigned_to"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// TaskToResponse converts Task model to response
func TaskToResponse(task *models.Task) TaskResponse {
	resp := TaskResponse{
		ID:             task.ID,
		Title:          task.Title,
		Description:    task.Description,
		TaskableType:   task.TaskableType,
		TaskableID:     task.TaskableID,
		StartDate:      task.StartDate,
		DueDate:        task.DueDate,
		CompletedAt:    task.CompletedAt,
		EstimatedHours: decimalToString(task.EstimatedHours),
		ActualHours:    decimalToString(task.ActualHours),
		Complexity:     task.Complexity,
		SortOrder:      task.SortOrder,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}

	if task.CurrentStatus != nil {
		resp.CurrentStatus = &StatusBrief{
			ID:       task.CurrentStatus.ID,
			Name:     task.CurrentStatus.Name,
			Category: task.CurrentStatus.Category,
		}
	}

	if task.Priority != nil {
		resp.Priority = &StatusBrief{
			ID:       task.Priority.ID,
			Name:     task.Priority.Name,
			Category: task.Priority.Category,
		}
	}

	if task.AssignedTo != nil {
		resp.AssignedTo = &IndividualBrief{
			ID:        task.AssignedTo.ID,
			FirstName: task.AssignedTo.FirstName,
			LastName:  task.AssignedTo.LastName,
			Email:     task.AssignedTo.Email,
		}
	}

	return resp
}
