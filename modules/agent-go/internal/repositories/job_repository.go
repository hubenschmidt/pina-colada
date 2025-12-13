package repositories

import (
	"errors"
	"time"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JobCreateInput contains data needed to create a job
type JobCreateInput struct {
	TenantID        *int64
	UserID          int64
	JobTitle        string
	Description     *string
	Source          string
	JobURL          *string
	SalaryRange     *string
	SalaryRangeID   *int64
	ResumeDate      *time.Time
	ProjectIDs      []int64
	AccountID       *int64
	CurrentStatusID *int64
}

// JobRepository handles job data access
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// FindAll returns paginated jobs with filtering
func (r *JobRepository) FindAll(params PaginationParams, tenantID *int64, projectID *int64) (*PaginatedResult[models.Job], error) {
	var jobs []models.Job
	var totalCount int64

	query := r.db.Model(&models.Job{}).
		Preload("Lead").
		Preload("Lead.CurrentStatus").
		Preload("Lead.Account").
		Preload("Lead.Account.Organizations").
		Preload("Lead.Account.Individuals").
		Preload("Lead.Projects").
		Preload("SalaryRangeRef").
		Joins(`JOIN "Lead" ON "Lead".id = "Job".id`)

	// Filter by tenant
	if tenantID != nil {
		query = query.Where(`"Lead".tenant_id = ?`, *tenantID)
	}

	// Filter by project
	if projectID != nil {
		query = query.Joins(`JOIN "Lead_Project" ON "Lead_Project".lead_id = "Lead".id`).
			Where(`"Lead_Project".project_id = ?`, *projectID)
	}

	// Apply search
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Joins(`LEFT JOIN "Account" ON "Account".id = "Lead".account_id`).
			Joins(`LEFT JOIN "Organization" ON "Organization".account_id = "Account".id`).
			Where(`LOWER("Organization".name) LIKE LOWER(?) OR LOWER("Job".job_title) LIKE LOWER(?)`,
				searchTerm, searchTerm)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Map order_by to actual column
	orderColumn := mapJobOrderColumn(params.OrderBy)
	query = ApplyPagination(query, PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		OrderBy:  orderColumn,
		Order:    params.Order,
	})

	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[models.Job]{
		Items:      jobs,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(totalCount, params.PageSize),
	}, nil
}

// FindByID returns a job by ID with all relationships
func (r *JobRepository) FindByID(id int64) (*models.Job, error) {
	var job models.Job
	err := r.db.
		Preload("Lead").
		Preload("Lead.CurrentStatus").
		Preload("Lead.Account").
		Preload("Lead.Account.Organizations").
		Preload("Lead.Account.Individuals").
		Preload("Lead.Projects").
		Preload("SalaryRangeRef").
		First(&job, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// Create creates a new job with its Deal and Lead parents
func (r *JobRepository) Create(input JobCreateInput) (int64, error) {
	var jobID int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		deal := &models.Deal{
			TenantID:  input.TenantID,
			Name:      input.JobTitle,
			CreatedBy: input.UserID,
			UpdatedBy: input.UserID,
		}
		if err := tx.Create(deal).Error; err != nil {
			return err
		}

		lead := &models.Lead{
			TenantID:        input.TenantID,
			DealID:          deal.ID,
			Type:            "Job",
			Source:          &input.Source,
			AccountID:       input.AccountID,
			CurrentStatusID: input.CurrentStatusID,
			CreatedBy:       input.UserID,
			UpdatedBy:       input.UserID,
		}
		if err := tx.Create(lead).Error; err != nil {
			return err
		}

		job := &models.Job{
			ID:            lead.ID,
			JobTitle:      input.JobTitle,
			Description:   input.Description,
			JobURL:        input.JobURL,
			SalaryRange:   input.SalaryRange,
			SalaryRangeID: input.SalaryRangeID,
			ResumeDate:    input.ResumeDate,
		}
		if err := tx.Create(job).Error; err != nil {
			return err
		}

		for _, pid := range input.ProjectIDs {
			lp := models.LeadProject{LeadID: lead.ID, ProjectID: pid}
			if err := tx.Create(&lp).Error; err != nil {
				return err
			}
		}

		jobID = job.ID
		return nil
	})
	return jobID, err
}

// Update updates an existing job by ID
func (r *JobRepository) Update(id int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Job{}).Where("id = ?", id).Omit(clause.Associations).Updates(updates).Error
}

// UpdateLead updates the lead associated with a job
func (r *JobRepository) UpdateLead(leadID int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Lead{}).Where("id = ?", leadID).Updates(updates).Error
}

// UpdateProjectAssociations replaces project associations for a lead
func (r *JobRepository) UpdateProjectAssociations(leadID int64, projectIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing
		if err := tx.Where("lead_id = ?", leadID).Delete(&models.LeadProject{}).Error; err != nil {
			return err
		}

		// Create new
		for _, pid := range projectIDs {
			lp := models.LeadProject{LeadID: leadID, ProjectID: pid}
			if err := tx.Create(&lp).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete deletes a job by ID
func (r *JobRepository) Delete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete job first (child)
		if err := tx.Delete(&models.Job{}, id).Error; err != nil {
			return err
		}
		// Delete lead (parent)
		if err := tx.Delete(&models.Lead{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// Count returns total job count
func (r *JobRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Job{}).Count(&count).Error
	return count, err
}

// GetRecentResumeDate returns the most recent resume date from all jobs
func (r *JobRepository) GetRecentResumeDate() (*time.Time, error) {
	var job models.Job
	err := r.db.
		Preload("Lead").
		Joins(`JOIN "Lead" ON "Lead".id = "Job".id`).
		Where(`"Job".resume_date IS NOT NULL`).
		Order(`"Lead".created_at DESC`).
		First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return job.ResumeDate, nil
}

var jobOrderColumns = map[string]string{
	"date":             `"Lead".created_at`,
	"application_date": `"Lead".created_at`,
	"updated_at":       `"Lead".updated_at`,
	"job_title":        `"Job".job_title`,
	"resume":           `"Job".resume_date`,
}

func mapJobOrderColumn(orderBy string) string {
	col, ok := jobOrderColumns[orderBy]
	if !ok {
		return `"Lead".updated_at`
	}
	return col
}

// FindAllStatuses returns all job-related statuses
func (r *JobRepository) FindAllStatuses() ([]models.Status, error) {
	var statuses []models.Status
	err := r.db.Where("category = ?", "job").Order("id ASC").Find(&statuses).Error
	return statuses, err
}

// FindJobsByStatusNames returns jobs filtered by status names
func (r *JobRepository) FindJobsByStatusNames(statusNames []string, tenantID *int64) ([]models.Job, error) {
	var jobs []models.Job

	query := r.db.Model(&models.Job{}).
		Preload("Lead").
		Preload("Lead.CurrentStatus").
		Preload("Lead.Account").
		Preload("Lead.Account.Organizations").
		Preload("Lead.Account.Individuals").
		Preload("Lead.Projects").
		Preload("SalaryRangeRef").
		Joins(`JOIN "Lead" ON "Lead".id = "Job".id`).
		Joins(`LEFT JOIN "Status" ON "Status".id = "Lead".current_status_id`)

	if tenantID != nil {
		query = query.Where(`"Lead".tenant_id = ?`, *tenantID)
	}

	if len(statusNames) > 0 {
		query = query.Where(`"Status".name IN ?`, statusNames)
	}

	query = query.Order(`"Lead".updated_at DESC`)

	err := query.Find(&jobs).Error
	return jobs, err
}

// UpdateJobStatus updates the status of a job's lead by status name
func (r *JobRepository) UpdateJobStatus(jobID int64, statusName string, userID int64) error {
	var status models.Status
	if err := r.db.Where("name = ? AND category = ?", statusName, "job").First(&status).Error; err != nil {
		return err
	}

	return r.db.Model(&models.Lead{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"current_status_id": status.ID,
		"updated_by":        userID,
	}).Error
}
