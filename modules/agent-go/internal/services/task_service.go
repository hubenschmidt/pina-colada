package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
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
func (s *TaskService) GetTasks(page, pageSize int, orderBy, order, search, scope string, tenantID *int64, taskableType *string, taskableID *int64, projectID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	// Only apply project filter when scope is "project"
	var effectiveProjectID *int64
	if scope == "project" && projectID != nil {
		effectiveProjectID = projectID
	}

	result, err := s.taskRepo.FindAll(params, tenantID, taskableType, taskableID, effectiveProjectID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.TaskResponse, len(result.Items))
	for i := range result.Items {
		entityInfo := s.resolveEntityInfo(result.Items[i].TaskableType, result.Items[i].TaskableID)
		items[i] = serializers.TaskToResponseWithEntity(&result.Items[i], entityInfo)
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

	entityInfo := s.resolveEntityInfo(task.TaskableType, task.TaskableID)
	resp := serializers.TaskToResponseWithEntity(task, entityInfo)
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

	repoInput := repositories.TaskCreateInput{
		TenantID:               tenantID,
		UserID:                 userID,
		Title:                  input.Title,
		Description:            input.Description,
		TaskableType:           input.TaskableType,
		TaskableID:             input.TaskableID,
		CurrentStatusID:        input.CurrentStatusID,
		PriorityID:             input.PriorityID,
		Complexity:             input.Complexity,
		AssignedToIndividualID: input.AssignedToIndividualID,
	}

	if input.SortOrder != nil {
		repoInput.SortOrder = *input.SortOrder
	}
	if input.StartDate != nil {
		repoInput.StartDate = parseDate(input.StartDate)
	}
	if input.DueDate != nil {
		repoInput.DueDate = parseDate(input.DueDate)
	}
	if input.EstimatedHours != nil {
		d := decimal.NewFromFloat(*input.EstimatedHours)
		repoInput.EstimatedHours = &d
	}
	if input.ActualHours != nil {
		d := decimal.NewFromFloat(*input.ActualHours)
		repoInput.ActualHours = &d
	}

	taskID, err := s.taskRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}

	return s.GetTask(taskID)
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
		entityInfo := s.resolveEntityInfo(task.TaskableType, task.TaskableID)
		resp := serializers.TaskToResponseWithEntity(task, entityInfo)
		return &resp, nil
	}

	if err := s.taskRepo.Update(id, updates); err != nil {
		return nil, err
	}

	return s.GetTask(id)
}

// GetTasksByEntity returns tasks for a specific entity
func (s *TaskService) GetTasksByEntity(entityType string, entityID int64, tenantID *int64) ([]serializers.TaskResponse, error) {
	tasks, err := s.taskRepo.FindByEntity(entityType, entityID, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]serializers.TaskResponse, len(tasks))
	for i := range tasks {
		entityInfo := s.resolveEntityInfo(tasks[i].TaskableType, tasks[i].TaskableID)
		result[i] = serializers.TaskToResponseWithEntity(&tasks[i], entityInfo)
	}
	return result, nil
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

// resolveEntityInfo resolves entity display info for a task
func (s *TaskService) resolveEntityInfo(taskableType *string, taskableID *int64) *serializers.TaskEntityInfo {
	if taskableType == nil || taskableID == nil {
		return nil
	}

	tType := *taskableType
	tID := *taskableID

	displayName := s.taskRepo.GetEntityDisplayName(tType, tID)
	url, leadType := s.buildEntityURL(tType, tID)

	return &serializers.TaskEntityInfo{
		Type:        &tType,
		ID:          &tID,
		DisplayName: displayName,
		URL:         url,
		LeadType:    leadType,
	}
}

// buildEntityURL builds URL and lead_type for an entity
func (s *TaskService) buildEntityURL(taskableType string, taskableID int64) (*string, *string) {
	urlMap := map[string]string{
		"Project":      fmt.Sprintf("/projects/%d", taskableID),
		"Deal":         fmt.Sprintf("/leads/deals/%d", taskableID),
		"Individual":   fmt.Sprintf("/accounts/individuals/%d", taskableID),
		"Organization": fmt.Sprintf("/accounts/organizations/%d", taskableID),
	}

	if url, ok := urlMap[taskableType]; ok {
		return &url, nil
	}

	if taskableType == "Lead" {
		return s.buildLeadURL(taskableID)
	}

	if taskableType == "Account" {
		return s.buildAccountURL(taskableID)
	}

	return nil, nil
}

func (s *TaskService) buildLeadURL(taskableID int64) (*string, *string) {
	leadType := s.taskRepo.GetLeadType(taskableID)
	if leadType == nil {
		return nil, nil
	}

	leadTypePlural := pluralizeLeadType(*leadType)
	url := fmt.Sprintf("/leads/%s/%d", leadTypePlural, taskableID)
	return &url, leadType
}

func pluralizeLeadType(leadType string) string {
	if leadType == "Opportunity" {
		return "opportunities"
	}
	return strings.ToLower(leadType) + "s"
}

func (s *TaskService) buildAccountURL(taskableID int64) (*string, *string) {
	entityType, entityID := s.taskRepo.GetAccountEntity(taskableID)
	if entityType == nil || entityID == nil {
		return nil, nil
	}

	pathMap := map[string]string{
		"Individual":   "individuals",
		"Organization": "organizations",
	}

	path, ok := pathMap[*entityType]
	if !ok {
		return nil, nil
	}

	url := fmt.Sprintf("/accounts/%s/%d", path, *entityID)
	return &url, nil
}
