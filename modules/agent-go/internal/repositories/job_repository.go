package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

// Create creates a new job with its Deal and Lead parents
func (r *JobRepository) Create(job *models.Job, lead *models.Lead, projectIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		deal := &models.Deal{
			TenantID:  lead.TenantID,
			Name:      lead.Title,
			CreatedBy: lead.CreatedBy,
			UpdatedBy: lead.UpdatedBy,
		}
		if err := tx.Create(deal).Error; err != nil {
			return err
		}

		lead.DealID = deal.ID
		if err := tx.Create(lead).Error; err != nil {
			return err
		}

		job.ID = lead.ID
		if err := tx.Create(job).Error; err != nil {
			return err
		}

		for _, pid := range projectIDs {
			lp := models.LeadProject{LeadID: lead.ID, ProjectID: pid}
			if err := tx.Create(&lp).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Update updates an existing job
func (r *JobRepository) Update(job *models.Job, updates map[string]interface{}) error {
	return r.db.Model(job).Omit(clause.Associations).Updates(updates).Error
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

func mapJobOrderColumn(orderBy string) string {
	switch orderBy {
	case "date", "application_date":
		return `"Lead".created_at`
	case "updated_at":
		return `"Lead".updated_at`
	case "job_title":
		return `"Job".job_title`
	case "resume":
		return `"Job".resume_date`
	default:
		return `"Lead".updated_at`
	}
}
