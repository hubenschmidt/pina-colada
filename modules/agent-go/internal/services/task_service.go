package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/shopspring/decimal"
)

var ErrTaskNotFound = errors.New("task not found")
var ErrTaskTitleRequired = errors.New("title is required")

// TaskService handles task business logic
type TaskService struct {
	taskRepo *repositories.TaskRepository
}

// NewTaskService creates a new task service
func NewTaskService(taskRepo *repositories.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

// GetTasks returns paginated tasks
func (s *TaskService) GetTasks(page, pageSize int, orderBy, order, search string, tenantID *int64, taskableType *string, taskableID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	result, err := s.taskRepo.FindAll(params, tenantID, taskableType, taskableID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.TaskResponse, len(result.Items))
	for i := range result.Items {
		items[i] = serializers.TaskToResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

// GetTask returns a task by ID
func (s *TaskService) GetTask(id int64) (*serializers.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}

	resp := serializers.TaskToResponse(task)
	return &resp, nil
}

// DeleteTask deletes a task
func (s *TaskService) DeleteTask(id int64) error {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	return s.taskRepo.Delete(id)
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(input schemas.TaskCreate, tenantID *int64, userID int64) (*serializers.TaskResponse, error) {
	if input.Title == "" {
		return nil, ErrTaskTitleRequired
	}

	task := &models.Task{
		TenantID:               tenantID,
		Title:                  input.Title,
		Description:            input.Description,
		TaskableType:           input.TaskableType,
		TaskableID:             input.TaskableID,
		CurrentStatusID:        input.CurrentStatusID,
		PriorityID:             input.PriorityID,
		Complexity:             input.Complexity,
		AssignedToIndividualID: input.AssignedToIndividualID,
		CreatedBy:              userID,
		UpdatedBy:              userID,
	}

	if input.SortOrder != nil {
		task.SortOrder = *input.SortOrder
	}

	if input.StartDate != nil {
		task.StartDate = parseDate(input.StartDate)
	}
	if input.DueDate != nil {
		task.DueDate = parseDate(input.DueDate)
	}
	if input.EstimatedHours != nil {
		d := decimal.NewFromFloat(*input.EstimatedHours)
		task.EstimatedHours = &d
	}
	if input.ActualHours != nil {
		d := decimal.NewFromFloat(*input.ActualHours)
		task.ActualHours = &d
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return s.GetTask(task.ID)
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(id int64, input schemas.TaskUpdate, userID int64) (*serializers.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	updates := buildTaskUpdates(input, userID)
	if len(updates) == 0 {
		resp := serializers.TaskToResponse(task)
		return &resp, nil
	}

	if err := s.taskRepo.Update(task, updates); err != nil {
		return nil, err
	}

	return s.GetTask(id)
}

func buildTaskUpdates(input schemas.TaskUpdate, userID int64) map[string]interface{} {
	updates := map[string]interface{}{
		"updated_by": userID,
	}

	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.TaskableType != nil {
		updates["taskable_type"] = *input.TaskableType
	}
	if input.TaskableID != nil {
		updates["taskable_id"] = *input.TaskableID
	}
	if input.CurrentStatusID != nil {
		updates["current_status_id"] = *input.CurrentStatusID
	}
	if input.PriorityID != nil {
		updates["priority_id"] = *input.PriorityID
	}
	if input.StartDate != nil {
		updates["start_date"] = parseDate(input.StartDate)
	}
	if input.DueDate != nil {
		updates["due_date"] = parseDate(input.DueDate)
	}
	if input.EstimatedHours != nil {
		updates["estimated_hours"] = decimal.NewFromFloat(*input.EstimatedHours)
	}
	if input.ActualHours != nil {
		updates["actual_hours"] = decimal.NewFromFloat(*input.ActualHours)
	}
	if input.Complexity != nil {
		updates["complexity"] = *input.Complexity
	}
	if input.SortOrder != nil {
		updates["sort_order"] = *input.SortOrder
	}
	if input.CompletedAt != nil {
		updates["completed_at"] = parseDate(input.CompletedAt)
	}
	if input.AssignedToIndividualID != nil {
		updates["assigned_to_individual_id"] = *input.AssignedToIndividualID
	}

	return updates
}

