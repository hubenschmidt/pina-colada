package repositories

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// TaskCreateInput contains data needed to create a task
type TaskCreateInput struct {
	TenantID               *int64
	UserID                 int64
	Title                  string
	Description            *string
	TaskableType           *string
	TaskableID             *int64
	CurrentStatusID        *int64
	PriorityID             *int64
	StartDate              *time.Time
	DueDate                *time.Time
	EstimatedHours         *decimal.Decimal
	ActualHours            *decimal.Decimal
	Complexity             *int16
	SortOrder              int
	AssignedToIndividualID *int64
}

// TaskRepository handles task data access
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// FindAll returns paginated tasks
func (r *TaskRepository) FindAll(params PaginationParams, tenantID *int64, taskableType *string, taskableID *int64, projectID *int64) (*PaginatedResult[models.Task], error) {
	var tasks []models.Task
	var totalCount int64

	query := r.db.Model(&models.Task{}).
		Preload("CurrentStatus").
		Preload("Priority").
		Preload("AssignedTo")

	if tenantID != nil {
		query = query.Where(`"Task".tenant_id = ?`, *tenantID)
	}

	if taskableType != nil && *taskableType != "" {
		query = query.Where(`"Task".taskable_type = ?`, *taskableType)
	}

	if taskableID != nil {
		query = query.Where(`"Task".taskable_id = ?`, *taskableID)
	}

	if projectID != nil {
		query = r.applyProjectFilter(query, *projectID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(`LOWER("Task".title) LIKE LOWER(?)`, searchTerm)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query = ApplyPagination(query, params)

	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[models.Task]{
		Items:      tasks,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// applyProjectFilter filters tasks for a project scope
func (r *TaskRepository) applyProjectFilter(query *gorm.DB, projectID int64) *gorm.DB {
	// Get Deal IDs for this project
	var dealIDs []int64
	r.db.Table(`"Deal"`).Select("id").Where("project_id = ?", projectID).Pluck("id", &dealIDs)

	// Get Lead IDs linked to this project
	var leadIDs []int64
	r.db.Table(`"Lead_Project"`).Select("lead_id").Where("project_id = ?", projectID).Pluck("lead_id", &leadIDs)

	// Get Account IDs linked to this project
	var accountIDs []int64
	r.db.Table(`"Account_Project"`).Select("account_id").Where("project_id = ?", projectID).Pluck("account_id", &accountIDs)

	// Build OR conditions for project scope
	conditions := r.db.Where(`"Task".taskable_type = ? AND "Task".taskable_id = ?`, "Project", projectID)

	if len(dealIDs) > 0 {
		conditions = conditions.Or(`"Task".taskable_type = ? AND "Task".taskable_id IN ?`, "Deal", dealIDs)
	}
	if len(leadIDs) > 0 {
		conditions = conditions.Or(`"Task".taskable_type = ? AND "Task".taskable_id IN ?`, "Lead", leadIDs)
	}
	if len(accountIDs) > 0 {
		conditions = conditions.Or(`"Task".taskable_type = ? AND "Task".taskable_id IN ?`, "Account", accountIDs)
	}

	return query.Where(conditions)
}

// FindByID returns a task by ID
func (r *TaskRepository) FindByID(id int64) (*models.Task, error) {
	var task models.Task
	err := r.db.
		Preload("CurrentStatus").
		Preload("Priority").
		Preload("AssignedTo").
		First(&task, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Create creates a new task
func (r *TaskRepository) Create(input TaskCreateInput) (int64, error) {
	task := &models.Task{
		TenantID:               input.TenantID,
		Title:                  input.Title,
		Description:            input.Description,
		TaskableType:           input.TaskableType,
		TaskableID:             input.TaskableID,
		CurrentStatusID:        input.CurrentStatusID,
		PriorityID:             input.PriorityID,
		StartDate:              input.StartDate,
		DueDate:                input.DueDate,
		EstimatedHours:         input.EstimatedHours,
		ActualHours:            input.ActualHours,
		Complexity:             input.Complexity,
		SortOrder:              input.SortOrder,
		AssignedToIndividualID: input.AssignedToIndividualID,
		CreatedBy:              input.UserID,
		UpdatedBy:              input.UserID,
	}
	err := r.db.Create(task).Error
	return task.ID, err
}

// Update updates a task by ID
func (r *TaskRepository) Update(id int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Task{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a task by ID
func (r *TaskRepository) Delete(id int64) error {
	return r.db.Delete(&models.Task{}, id).Error
}

// FindByEntity returns tasks for a specific entity
func (r *TaskRepository) FindByEntity(entityType string, entityID int64, tenantID *int64) ([]models.Task, error) {
	var tasks []models.Task
	query := r.db.
		Preload("CurrentStatus").
		Preload("Priority").
		Preload("AssignedTo").
		Where("taskable_type = ? AND taskable_id = ?", entityType, entityID)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	err := query.Order("sort_order ASC, created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetEntityDisplayName returns display name for an entity
func (r *TaskRepository) GetEntityDisplayName(entityType string, entityID int64) *string {
	if entityType == "Lead" {
		return r.getLeadDisplayName(entityID)
	}

	tableSelect := map[string]string{
		"Project":      "name",
		"Deal":         "name",
		"Account":      "name",
		"Individual":   "first_name || ' ' || last_name",
		"Organization": "name",
	}

	selectField, ok := tableSelect[entityType]
	if !ok {
		return nil
	}

	var name string
	err := r.db.Table(`"`+entityType+`"`).Select(selectField).Where("id = ?", entityID).Scan(&name).Error
	if err != nil || name == "" {
		return nil
	}
	return &name
}

// getLeadDisplayName gets display name from Lead subtype tables (Job, Opportunity, Partnership)
func (r *TaskRepository) getLeadDisplayName(leadID int64) *string {
	var name string

	// Try Job
	if err := r.db.Table(`"Job"`).Select("job_title").Where("id = ?", leadID).Scan(&name).Error; err == nil && name != "" {
		return &name
	}

	// Try Opportunity
	if err := r.db.Table(`"Opportunity"`).Select("opportunity_name").Where("id = ?", leadID).Scan(&name).Error; err == nil && name != "" {
		return &name
	}

	// Try Partnership
	if err := r.db.Table(`"Partnership"`).Select("partnership_name").Where("id = ?", leadID).Scan(&name).Error; err == nil && name != "" {
		return &name
	}

	return nil
}

// GetLeadType returns the type of a lead
func (r *TaskRepository) GetLeadType(leadID int64) *string {
	var leadType string
	err := r.db.Table(`"Lead"`).Select("type").Where("id = ?", leadID).Scan(&leadType).Error
	if err != nil || leadType == "" {
		return nil
	}
	return &leadType
}

// GetAccountEntity returns the Individual or Organization linked to an Account
func (r *TaskRepository) GetAccountEntity(accountID int64) (*string, *int64) {
	// Check for Individual
	var indID int64
	err := r.db.Table(`"Individual"`).Select("id").Where("account_id = ?", accountID).Limit(1).Scan(&indID).Error
	if err == nil && indID > 0 {
		entityType := "Individual"
		return &entityType, &indID
	}

	// Check for Organization
	var orgID int64
	err = r.db.Table(`"Organization"`).Select("id").Where("account_id = ?", accountID).Limit(1).Scan(&orgID).Error
	if err == nil && orgID > 0 {
		entityType := "Organization"
		return &entityType, &orgID
	}

	return nil, nil
}
