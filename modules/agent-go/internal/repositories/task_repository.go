package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// TaskRepository handles task data access
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// FindAll returns paginated tasks
func (r *TaskRepository) FindAll(params PaginationParams, tenantID *int64, taskableType *string, taskableID *int64) (*PaginatedResult[models.Task], error) {
	var tasks []models.Task
	var totalCount int64

	query := r.db.Model(&models.Task{}).
		Preload("CurrentStatus").
		Preload("Priority").
		Preload("AssignedTo")

	if tenantID != nil {
		query = query.Where("tasks.tenant_id = ?", *tenantID)
	}

	if taskableType != nil && *taskableType != "" {
		query = query.Where("tasks.taskable_type = ?", *taskableType)
	}

	if taskableID != nil {
		query = query.Where("tasks.taskable_id = ?", *taskableID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("LOWER(tasks.title) LIKE LOWER(?)", searchTerm)
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

// FindByID returns a task by ID
func (r *TaskRepository) FindByID(id int64) (*models.Task, error) {
	var task models.Task
	err := r.db.
		Preload("CurrentStatus").
		Preload("Priority").
		Preload("AssignedTo").
		First(&task, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// Create creates a new task
func (r *TaskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

// Update updates a task
func (r *TaskRepository) Update(task *models.Task, updates map[string]interface{}) error {
	return r.db.Model(task).Updates(updates).Error
}

// Delete deletes a task by ID
func (r *TaskRepository) Delete(id int64) error {
	return r.db.Delete(&models.Task{}, id).Error
}
