package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// ReportRepository handles report data access
type ReportRepository struct {
	db *gorm.DB
}

// NewReportRepository creates a new report repository
func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// SavedReportWithProjects represents a saved report with its project associations
type SavedReportWithProjects struct {
	models.SavedReport
	ProjectIDs []int64 `gorm:"-"`
}

// FindSavedReports returns saved reports for a tenant with pagination
func (r *ReportRepository) FindSavedReports(tenantID int64, projectID *int64, includeGlobal bool, search string, page, limit int, sortBy, order string) ([]models.SavedReport, int64, error) {
	var reports []models.SavedReport
	var total int64

	query := r.db.Model(&models.SavedReport{}).Where("tenant_id = ?", tenantID)

	if projectID != nil {
		subquery := r.db.Table(`"Saved_Report_Project"`).Select("saved_report_id").Where("project_id = ?", *projectID)
		if includeGlobal {
			// Include reports with this project OR reports with no project associations
			noProjectSubquery := r.db.Table(`"Saved_Report_Project"`).Select("saved_report_id")
			query = query.Where("id IN (?) OR id NOT IN (?)", subquery, noProjectSubquery)
		} else {
			query = query.Where("id IN (?)", subquery)
		}
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if sortBy == "" {
		sortBy = "updated_at"
	}
	if order == "" {
		order = "DESC"
	}
	query = query.Order(sortBy + " " + order)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if page > 0 && limit > 0 {
		query = query.Offset((page - 1) * limit)
	}

	err := query.Find(&reports).Error
	return reports, total, err
}

// FindSavedReportByID returns a saved report by ID
func (r *ReportRepository) FindSavedReportByID(reportID int64, tenantID int64) (*models.SavedReport, error) {
	var report models.SavedReport
	err := r.db.Where("id = ? AND tenant_id = ?", reportID, tenantID).First(&report).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

// GetProjectIDsForReport returns project IDs associated with a saved report
func (r *ReportRepository) GetProjectIDsForReport(reportID int64) ([]int64, error) {
	var projectIDs []int64
	err := r.db.Table(`"Saved_Report_Project"`).
		Where("saved_report_id = ?", reportID).
		Pluck("project_id", &projectIDs).Error
	return projectIDs, err
}

// CreateSavedReport creates a new saved report
func (r *ReportRepository) CreateSavedReport(report *models.SavedReport) error {
	return r.db.Create(report).Error
}

// UpdateSavedReport updates a saved report
func (r *ReportRepository) UpdateSavedReport(reportID int64, updates map[string]interface{}) error {
	return r.db.Model(&models.SavedReport{}).Where("id = ?", reportID).Updates(updates).Error
}

// DeleteSavedReport deletes a saved report and its project associations
func (r *ReportRepository) DeleteSavedReport(reportID int64) error {
	tx := r.db.Begin()

	// Delete project associations
	if err := tx.Table(`"Saved_Report_Project"`).Where("saved_report_id = ?", reportID).Delete(nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the report
	if err := tx.Delete(&models.SavedReport{}, reportID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// SetProjectsForReport sets the project associations for a saved report
func (r *ReportRepository) SetProjectsForReport(reportID int64, projectIDs []int64) error {
	tx := r.db.Begin()

	// Delete existing associations
	if err := tx.Table(`"Saved_Report_Project"`).Where("saved_report_id = ?", reportID).Delete(nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Insert new associations
	for _, pid := range projectIDs {
		assoc := models.SavedReportProject{SavedReportID: reportID, ProjectID: pid}
		if err := tx.Create(&assoc).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetLeadPipelineData returns data for the lead pipeline canned report
func (r *ReportRepository) GetLeadPipelineData(tenantID int64, dateFrom, dateTo *string, projectID *int64) ([]map[string]interface{}, error) {
	query := r.db.Table(`"Lead"`).
		Select(`"Lead".id, "Lead".title, "Lead".source, "Lead".type, "Lead".created_at, "Account".name as account_name`).
		Joins(`LEFT JOIN "Account" ON "Lead".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID)

	if dateFrom != nil && *dateFrom != "" {
		query = query.Where(`"Lead".created_at >= ?`, *dateFrom)
	}
	if dateTo != nil && *dateTo != "" {
		query = query.Where(`"Lead".created_at <= ?`, *dateTo)
	}
	if projectID != nil {
		query = query.Joins(`INNER JOIN "Lead_Project" ON "Lead".id = "Lead_Project".lead_id`).
			Where(`"Lead_Project".project_id = ?`, *projectID)
	}

	query = query.Order(`"Lead".created_at DESC`)

	var results []map[string]interface{}
	err := query.Find(&results).Error
	return results, err
}

// GetAccountOverviewData returns data for the account overview canned report
func (r *ReportRepository) GetAccountOverviewData(tenantID int64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Count organizations
	var orgCount int64
	r.db.Table(`"Organization"`).
		Joins(`JOIN "Account" ON "Organization".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&orgCount)
	result["organizations"] = orgCount

	// Count individuals
	var indCount int64
	r.db.Table(`"Individual"`).
		Joins(`JOIN "Account" ON "Individual".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&indCount)
	result["individuals"] = indCount

	// Count leads
	var leadCount int64
	r.db.Table(`"Lead"`).
		Joins(`JOIN "Account" ON "Lead".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&leadCount)
	result["leads"] = leadCount

	// Count contacts
	var contactCount int64
	r.db.Table(`"Contact"`).
		Joins(`JOIN "Account" ON "Contact".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&contactCount)
	result["contacts"] = contactCount

	return result, nil
}

// GetContactCoverageData returns data for the contact coverage canned report
func (r *ReportRepository) GetContactCoverageData(tenantID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Table(`"Organization"`).
		Select(`"Organization".id, "Organization".name, COUNT("Contact".id) as contact_count`).
		Joins(`JOIN "Account" ON "Organization".account_id = "Account".id`).
		Joins(`LEFT JOIN "Contact" ON "Contact".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Group(`"Organization".id, "Organization".name`).
		Order(`contact_count DESC`).
		Find(&results).Error
	return results, err
}

// GetNotesActivityData returns data for the notes activity canned report
func (r *ReportRepository) GetNotesActivityData(tenantID int64, projectID *int64) ([]map[string]interface{}, error) {
	query := r.db.Table(`"Note"`).
		Select(`"Note".id, "Note".entity_type, "Note".entity_id, "Note".content, "Note".created_at, "User".email as created_by_email`).
		Joins(`LEFT JOIN "User" ON "Note".created_by = "User".id`).
		Where(`"Note".tenant_id = ?`, tenantID).
		Order(`"Note".created_at DESC`).
		Limit(100)

	var results []map[string]interface{}
	err := query.Find(&results).Error
	return results, err
}

// GetUserAuditData returns data for the user audit canned report
func (r *ReportRepository) GetUserAuditData(tenantID int64, userID *int64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	userQuery := r.db.Table(`"User"`).Where("tenant_id = ?", tenantID)
	if userID != nil {
		userQuery = userQuery.Where("id = ?", *userID)
	}

	var users []map[string]interface{}
	userQuery.Select("id, email, first_name, last_name, created_at, last_login_at").Find(&users)
	result["users"] = users

	// Get activity counts
	var activityCounts []map[string]interface{}
	r.db.Table(`"Note"`).
		Select(`created_by as user_id, COUNT(*) as note_count`).
		Where("tenant_id = ?", tenantID).
		Group("created_by").
		Find(&activityCounts)
	result["activity_counts"] = activityCounts

	return result, nil
}

// ExecuteCustomQuery executes a custom report query
func (r *ReportRepository) ExecuteCustomQuery(tenantID int64, primaryEntity string, filters []map[string]interface{}, limit, offset int, projectID *int64) ([]map[string]interface{}, int64, error) {
	tableName := getTableName(primaryEntity)
	if tableName == "" {
		return nil, 0, nil
	}

	query := r.db.Table(tableName)

	// Join with Account for tenant filtering (except for notes which have tenant_id directly)
	if primaryEntity == "notes" {
		query = query.Where(`"Note".tenant_id = ?`, tenantID)
	} else {
		query = query.Joins(`JOIN "Account" ON "` + tableName + `".account_id = "Account".id`).
			Where(`"Account".tenant_id = ?`, tenantID)
	}

	// Apply filters
	for _, f := range filters {
		field, ok := f["field"].(string)
		if !ok {
			continue
		}
		operator, ok := f["operator"].(string)
		if !ok {
			continue
		}
		value := f["value"]

		query = applyFilter(query, tableName, field, operator, value)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var results []map[string]interface{}
	err := query.Find(&results).Error
	return results, total, err
}

func getTableName(entity string) string {
	switch entity {
	case "organizations":
		return "Organization"
	case "individuals":
		return "Individual"
	case "contacts":
		return "Contact"
	case "leads":
		return "Lead"
	case "notes":
		return "Note"
	default:
		return ""
	}
}

func applyFilter(query *gorm.DB, table, field, operator string, value interface{}) *gorm.DB {
	fullField := `"` + table + `"."` + field + `"`

	switch operator {
	case "eq":
		return query.Where(fullField+" = ?", value)
	case "neq":
		return query.Where(fullField+" != ?", value)
	case "gt":
		return query.Where(fullField+" > ?", value)
	case "gte":
		return query.Where(fullField+" >= ?", value)
	case "lt":
		return query.Where(fullField+" < ?", value)
	case "lte":
		return query.Where(fullField+" <= ?", value)
	case "contains":
		return query.Where("LOWER("+fullField+") LIKE LOWER(?)", "%"+value.(string)+"%")
	case "starts_with":
		return query.Where("LOWER("+fullField+") LIKE LOWER(?)", value.(string)+"%")
	case "is_null":
		return query.Where(fullField + " IS NULL")
	case "is_not_null":
		return query.Where(fullField + " IS NOT NULL")
	case "in":
		return query.Where(fullField+" IN ?", value)
	default:
		return query
	}
}
