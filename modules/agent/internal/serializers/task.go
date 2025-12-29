package serializers

import (
	"time"

	"agent/internal/models"
	"github.com/shopspring/decimal"
)

func decimalToString(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	s := d.String()
	return &s
}

// TaskStatusBrief represents status/priority for task
type TaskStatusBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TaskEntityInfo represents the linked entity for a task
type TaskEntityInfo struct {
	Type        *string `json:"type"`
	ID          *int64  `json:"id"`
	DisplayName *string `json:"display_name"`
	URL         *string `json:"url"`
	LeadType    *string `json:"lead_type"`
}

// TaskResponse represents a task
type TaskResponse struct {
	ID             int64            `json:"id"`
	Title          string           `json:"title"`
	Description    *string          `json:"description"`
	Status         *TaskStatusBrief `json:"status"`
	Priority       *TaskStatusBrief `json:"priority"`
	Entity         TaskEntityInfo   `json:"entity"`
	StartDate      *string          `json:"start_date"`
	DueDate        *string          `json:"due_date"`
	CompletedAt    *string          `json:"completed_at"`
	EstimatedHours *float64         `json:"estimated_hours"`
	ActualHours    *float64         `json:"actual_hours"`
	Complexity     *int16           `json:"complexity"`
	SortOrder      int              `json:"sort_order"`
	AssignedTo     *IndividualBrief `json:"assigned_to"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func formatDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func decimalToFloat(d *decimal.Decimal) *float64 {
	if d == nil {
		return nil
	}
	f, _ := d.Float64()
	return &f
}

// TaskToResponse converts Task model to response
func TaskToResponse(task *models.Task) TaskResponse {
	return TaskToResponseWithEntity(task, nil)
}

// TaskToResponseWithEntity converts Task model to response with entity info
func TaskToResponseWithEntity(task *models.Task, entityInfo *TaskEntityInfo) TaskResponse {
	resp := TaskResponse{
		ID:             task.ID,
		Title:          task.Title,
		Description:    task.Description,
		StartDate:      formatDate(task.StartDate),
		DueDate:        formatDate(task.DueDate),
		CompletedAt:    nil,
		EstimatedHours: decimalToFloat(task.EstimatedHours),
		ActualHours:    decimalToFloat(task.ActualHours),
		Complexity:     task.Complexity,
		SortOrder:      task.SortOrder,
		CreatedAt:      formatDateTime(task.CreatedAt),
		UpdatedAt:      formatDateTime(task.UpdatedAt),
		Entity:         TaskEntityInfo{},
	}

	if entityInfo != nil {
		resp.Entity = *entityInfo
	}

	if task.CompletedAt != nil {
		s := task.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &s
	}

	if task.CurrentStatus != nil {
		resp.Status = &TaskStatusBrief{
			ID:   task.CurrentStatus.ID,
			Name: task.CurrentStatus.Name,
		}
	}

	if task.Priority != nil {
		resp.Priority = &TaskStatusBrief{
			ID:   task.Priority.ID,
			Name: task.Priority.Name,
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
