package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

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
		return errors.New("task not found")
	}

	return s.taskRepo.Delete(id)
}
