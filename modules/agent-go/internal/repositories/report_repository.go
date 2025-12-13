package repositories

import (
	"errors"
	"fmt"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
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

// SavedReportCreateInput represents input for creating a saved report
type SavedReportCreateInput struct {
	TenantID        int64
	Name            string
	Description     *string
	QueryDefinition []byte
	CreatedBy       *int64
}

// FindSavedReports returns saved reports for a tenant with pagination
func (r *ReportRepository) FindSavedReports(tenantID int64, projectID *int64, includeGlobal bool, search string, page, limit int, sortBy, order string) ([]models.SavedReport, int64, error) {
	var reports []models.SavedReport
	var total int64

	query := r.db.Model(&models.SavedReport{}).Where("tenant_id = ?", tenantID)

	query = r.applyProjectFilter(query, projectID, includeGlobal)

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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
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
func (r *ReportRepository) CreateSavedReport(input SavedReportCreateInput) (int64, error) {
	report := &models.SavedReport{
		TenantID:        input.TenantID,
		Name:            input.Name,
		Description:     input.Description,
		QueryDefinition: input.QueryDefinition,
		CreatedBy:       input.CreatedBy,
	}
	if err := r.db.Create(report).Error; err != nil {
		return 0, err
	}
	return report.ID, nil
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
		Select(`"Lead".id, "Lead".source, "Lead".type, "Lead".created_at, "Account".name as account_name`).
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
	result["total_organizations"] = orgCount

	// Count individuals
	var indCount int64
	r.db.Table(`"Individual"`).
		Joins(`JOIN "Account" ON "Individual".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&indCount)
	result["total_individuals"] = indCount

	// Count leads
	var leadCount int64
	r.db.Table(`"Lead"`).
		Joins(`JOIN "Account" ON "Lead".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&leadCount)
	result["total_leads"] = leadCount

	// Count contacts
	var contactCount int64
	r.db.Table(`"Contact"`).
		Joins(`JOIN "Account" ON "Contact".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&contactCount)
	result["total_contacts"] = contactCount

	// Organizations by country
	var countryResults []struct {
		Country string
		Count   int64
	}
	r.db.Table(`"Organization"`).
		Select(`COALESCE("Organization".headquarters_country, 'Unknown') as country, COUNT(*) as count`).
		Joins(`JOIN "Account" ON "Organization".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Group("country").
		Find(&countryResults)

	byCountry := make(map[string]int64)
	for _, row := range countryResults {
		byCountry[row.Country] = row.Count
	}
	result["organizations_by_country"] = byCountry

	// Organizations by type
	var typeResults []struct {
		Type  string
		Count int64
	}
	r.db.Table(`"Organization"`).
		Select(`COALESCE("Organization".company_type, 'Unknown') as type, COUNT(*) as count`).
		Joins(`JOIN "Account" ON "Organization".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Group("type").
		Find(&typeResults)

	byType := make(map[string]int64)
	for _, row := range typeResults {
		byType[row.Type] = row.Count
	}
	result["organizations_by_type"] = byType

	return result, nil
}

// GetContactCoverageData returns data for the contact coverage canned report
func (r *ReportRepository) GetContactCoverageData(tenantID int64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Get coverage by org
	var coverageByOrg []map[string]interface{}
	r.db.Table(`"Organization"`).
		Select(`"Organization".id as organization_id, "Organization".name as organization_name, COUNT(DISTINCT "Contact".id) as contact_count`).
		Joins(`JOIN "Account" ON "Organization".account_id = "Account".id`).
		Joins(`LEFT JOIN "Contact_Account" ON "Contact_Account".account_id = "Account".id`).
		Joins(`LEFT JOIN "Contact" ON "Contact".id = "Contact_Account".contact_id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Group(`"Organization".id, "Organization".name`).
		Order(`contact_count DESC`).
		Find(&coverageByOrg)
	result["coverage_by_org"] = coverageByOrg

	// Total organizations
	var totalOrgs int64
	r.db.Table(`"Organization"`).
		Joins(`JOIN "Account" ON "Organization".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&totalOrgs)
	result["total_organizations"] = totalOrgs

	// Total contacts
	var totalContacts int64
	r.db.Table(`"Contact"`).
		Joins(`JOIN "Contact_Account" ON "Contact".id = "Contact_Account".contact_id`).
		Joins(`JOIN "Account" ON "Contact_Account".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID).
		Count(&totalContacts)
	result["total_contacts"] = totalContacts

	// Average contacts per org
	avgContactsPerOrg := float64(0)
	if totalOrgs > 0 {
		avgContactsPerOrg = float64(totalContacts) / float64(totalOrgs)
	}
	result["average_contacts_per_org"] = fmt.Sprintf("%.1f", avgContactsPerOrg)

	// Decision maker count
	var decisionMakerCount int64
	r.db.Table(`"Contact"`).
		Joins(`JOIN "Contact_Account" ON "Contact".id = "Contact_Account".contact_id`).
		Joins(`JOIN "Account" ON "Contact_Account".account_id = "Account".id`).
		Where(`"Account".tenant_id = ? AND "Contact".is_primary = true`, tenantID).
		Count(&decisionMakerCount)
	result["decision_maker_count"] = decisionMakerCount

	// Organizations with zero contacts
	var orgsWithZeroContacts int64
	for _, org := range coverageByOrg {
		count, ok := org["contact_count"].(int64)
		if ok && count == 0 {
			orgsWithZeroContacts++
		}
	}
	result["organizations_with_zero_contacts"] = orgsWithZeroContacts

	// Decision maker ratio
	decisionMakerRatio := float64(0)
	if totalContacts > 0 {
		decisionMakerRatio = float64(decisionMakerCount) / float64(totalContacts)
	}
	result["decision_maker_ratio"] = decisionMakerRatio

	return result, nil
}

// GetNotesActivityData returns data for the notes activity canned report
func (r *ReportRepository) GetNotesActivityData(tenantID int64, projectID *int64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Total notes
	var totalNotes int64
	r.db.Table(`"Note"`).
		Where(`"Note".tenant_id = ?`, tenantID).
		Count(&totalNotes)
	result["total_notes"] = totalNotes

	// Notes by entity type
	var byEntityType []struct {
		EntityType string
		Count      int64
	}
	r.db.Table(`"Note"`).
		Select(`entity_type, COUNT(*) as count`).
		Where(`tenant_id = ?`, tenantID).
		Group("entity_type").
		Find(&byEntityType)

	byTypeMap := make(map[string]int64)
	for _, row := range byEntityType {
		byTypeMap[row.EntityType] = row.Count
	}
	result["by_entity_type"] = byTypeMap

	// Unique entities with notes by type
	var entitiesWithNotes []struct {
		EntityType string
		Count      int64
	}
	r.db.Table(`"Note"`).
		Select(`entity_type, COUNT(DISTINCT entity_id) as count`).
		Where(`tenant_id = ?`, tenantID).
		Group("entity_type").
		Find(&entitiesWithNotes)

	entitiesMap := make(map[string]int64)
	for _, row := range entitiesWithNotes {
		entitiesMap[row.EntityType] = row.Count
	}
	result["entities_with_notes"] = entitiesMap

	// Recent notes
	var recentNotes []map[string]interface{}
	r.db.Table(`"Note"`).
		Select(`"Note".id, "Note".entity_type, "Note".entity_id, "Note".content, "Note".created_at, "User".email as created_by_email`).
		Joins(`LEFT JOIN "User" ON "Note".created_by = "User".id`).
		Where(`"Note".tenant_id = ?`, tenantID).
		Order(`"Note".created_at DESC`).
		Limit(100).
		Find(&recentNotes)
	result["recent_notes"] = recentNotes

	return result, nil
}

// getAuditTableCounts returns created and updated counts for a table using guard clauses
func (r *ReportRepository) getAuditTableCounts(table string, tenantID int64, userID *int64) (int64, int64) {
	// Contact joins through many-to-many, need DISTINCT to avoid duplicates
	if table == "Contact" {
		return r.getContactAuditCounts(tenantID, userID)
	}

	auditTable := r.getAuditTableName(table)
	createdCount := r.countAuditRecords(table, auditTable, tenantID, userID, "created_by")
	updatedCount := r.countAuditRecords(table, auditTable, tenantID, userID, "updated_by")

	return createdCount, updatedCount
}

// getAuditTableName returns the table containing audit columns (Asset for Document, else same table)
func (r *ReportRepository) getAuditTableName(table string) string {
	if table == "Document" {
		return "Asset"
	}
	return table
}

// getContactAuditCounts returns counts for Contact table with DISTINCT to handle many-to-many
func (r *ReportRepository) getContactAuditCounts(tenantID int64, userID *int64) (int64, int64) {
	var count int64
	query := r.db.Table(`"Contact"`).
		Select(`COUNT(DISTINCT "Contact".id)`).
		Joins(`JOIN "Contact_Account" ON "Contact".id = "Contact_Account".contact_id`).
		Joins(`JOIN "Account" ON "Contact_Account".account_id = "Account".id`).
		Where(`"Account".tenant_id = ?`, tenantID)
	query = r.applyUserFilter(query, "Contact", userID, "created_by")
	query.Scan(&count)
	return count, count
}

// countAuditRecords counts records for a table filtered by tenant and optionally user
func (r *ReportRepository) countAuditRecords(table, auditTable string, tenantID int64, userID *int64, auditCol string) int64 {
	var count int64
	query := r.applyAuditTenantFilter(r.db.Table(fmt.Sprintf(`"%s"`, table)), table, tenantID)
	query = r.applyUserFilter(query, auditTable, userID, auditCol)
	query.Count(&count)
	return count
}

// applyUserFilter adds user filter to query if userID is provided
func (r *ReportRepository) applyUserFilter(query *gorm.DB, table string, userID *int64, column string) *gorm.DB {
	if userID == nil {
		return query
	}
	return query.Where(fmt.Sprintf(`"%s".%s = ?`, table, column), *userID)
}

// applyAuditTenantFilter applies tenant filtering based on table type using guard clauses
func (r *ReportRepository) applyAuditTenantFilter(query *gorm.DB, table string, tenantID int64) *gorm.DB {
	// Tables with direct tenant_id column
	directTenantTables := map[string]bool{
		"Account": true, "Asset": true, "Comment": true, "Deal": true,
		"Lead": true, "Note": true, "Project": true, "Task": true,
	}
	if directTenantTables[table] {
		return query.Where("tenant_id = ?", tenantID)
	}

	// Tables needing Account join (have account_id but no tenant_id)
	if table == "Individual" || table == "Organization" {
		return query.Joins(fmt.Sprintf(`JOIN "Account" ON "%s".account_id = "Account".id`, table)).
			Where(`"Account".tenant_id = ?`, tenantID)
	}

	// Account_Relationship joins through from_account_id
	if table == "Account_Relationship" {
		return query.Joins(`JOIN "Account" ON "Account_Relationship".from_account_id = "Account".id`).
			Where(`"Account".tenant_id = ?`, tenantID)
	}

	// Document inherits from Asset - join Asset for tenant_id
	if table == "Document" {
		return query.Joins(`JOIN "Asset" ON "Document".id = "Asset".id`).
			Where(`"Asset".tenant_id = ?`, tenantID)
	}

	// Contact needs Contact_Account join
	if table == "Contact" {
		return query.Joins(`JOIN "Contact_Account" ON "Contact".id = "Contact_Account".contact_id`).
			Joins(`JOIN "Account" ON "Contact_Account".account_id = "Account".id`).
			Where(`"Account".tenant_id = ?`, tenantID)
	}

	return query
}

// GetUserAuditData returns data for the user audit canned report
func (r *ReportRepository) GetUserAuditData(tenantID int64, userID *int64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	r.loadAuditUserInfo(result, tenantID, userID)

	// Tables to audit with created_by/updated_by columns (13 total, Document inherits from Asset)
	tables := []string{"Account", "Account_Relationship", "Asset", "Comment", "Contact", "Deal", "Document", "Individual", "Lead", "Note", "Organization", "Project", "Task"}

	byTable := make([]map[string]interface{}, 0)
	var totalCreated int64
	var totalUpdated int64

	for _, table := range tables {
		createdCount, updatedCount := r.getAuditTableCounts(table, tenantID, userID)

		if createdCount > 0 || updatedCount > 0 {
			byTable = append(byTable, map[string]interface{}{
				"table":         table,
				"created_count": createdCount,
				"updated_count": updatedCount,
			})
		}

		totalCreated += createdCount
		totalUpdated += updatedCount
	}

	result["by_table"] = byTable
	result["total_created"] = totalCreated
	result["total_updated"] = totalUpdated

	// Recent activity - query each audit table for recent updates
	recentActivity := make([]map[string]interface{}, 0)
	limitPerTable := 5

	for _, table := range tables {
		auditTable := table
		if table == "Document" {
			auditTable = "Asset"
		}

		// Build SELECT with appropriate display_name column
		selectCols := r.getAuditSelectColumns(table, auditTable)
		query := r.db.Table(fmt.Sprintf(`"%s"`, table)).Select(selectCols)
		query = r.applyAuditTenantFilter(query, table, tenantID)
		query = query.Where(fmt.Sprintf(`"%s".updated_by IS NOT NULL`, auditTable))

		if userID != nil {
			query = query.Where(fmt.Sprintf(`"%s".updated_by = ?`, auditTable), *userID)
		}

		query = query.Order(fmt.Sprintf(`"%s".updated_at DESC`, auditTable)).Limit(limitPerTable)

		var rows []map[string]interface{}
		query.Find(&rows)

		for _, row := range rows {
			recentActivity = append(recentActivity, map[string]interface{}{
				"table":        table,
				"id":           row["id"],
				"display_name": row["display_name"],
				"updated_by":   row["updated_by"],
				"updated_at":   row["updated_at"],
			})
		}
	}

	result["recent_activity"] = recentActivity

	return result, nil
}

// getAuditSelectColumns returns the SELECT columns for a table's recent activity query
func (r *ReportRepository) getAuditSelectColumns(table, auditTable string) string {
	baseSelect := fmt.Sprintf(`"%s".id, "%s".updated_by, "%s".updated_at`, table, auditTable, auditTable)

	// Tables with name column
	if table == "Account" || table == "Organization" || table == "Project" {
		return baseSelect + fmt.Sprintf(`, "%s".name as display_name`, table)
	}

	// Tables with title column
	if table == "Deal" || table == "Task" || table == "Lead" {
		return baseSelect + fmt.Sprintf(`, "%s".title as display_name`, table)
	}

	// Tables with first_name/last_name
	if table == "Individual" || table == "Contact" {
		return baseSelect + fmt.Sprintf(`, CONCAT("%s".first_name, ' ', "%s".last_name) as display_name`, table, table)
	}

	// Asset/Document with filename
	if table == "Asset" || table == "Document" {
		return baseSelect + `, "Asset".filename as display_name`
	}

	// Note/Comment with content
	if table == "Note" || table == "Comment" {
		return baseSelect + fmt.Sprintf(`, SUBSTRING("%s".content, 1, 50) as display_name`, table)
	}

	// Account_Relationship - use notes or fallback
	if table == "Account_Relationship" {
		return baseSelect + fmt.Sprintf(`, COALESCE("%s".notes, CONCAT('#', "%s".id)) as display_name`, table, table)
	}

	return baseSelect + fmt.Sprintf(`, CONCAT('#', "%s".id) as display_name`, table)
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
	}
	if primaryEntity != "notes" {
		query = query.Joins(`JOIN "Account" ON "` + tableName + `".account_id = "Account".id`).
			Where(`"Account".tenant_id = ?`, tenantID)
	}

	// Apply filters
	for _, f := range filters {
		field, fieldOk := f["field"].(string)
		operator, opOk := f["operator"].(string)
		if fieldOk && opOk {
			query = applyFilter(query, tableName, field, operator, f["value"])
		}
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

var entityTableNames = map[string]string{
	"organizations": "Organization",
	"individuals":   "Individual",
	"contacts":      "Contact",
	"leads":         "Lead",
	"notes":         "Note",
}

func getTableName(entity string) string {
	return entityTableNames[entity]
}

type filterApplier func(query *gorm.DB, fullField string, value interface{}) *gorm.DB

var filterOperators = map[string]filterApplier{
	"eq":          func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f+" = ?", v) },
	"neq":         func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f+" != ?", v) },
	"gt":          func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f+" > ?", v) },
	"gte":         func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f+" >= ?", v) },
	"lt":          func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f+" < ?", v) },
	"lte":         func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f+" <= ?", v) },
	"contains":    func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where("LOWER("+f+") LIKE LOWER(?)", "%"+v.(string)+"%") },
	"starts_with": func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where("LOWER("+f+") LIKE LOWER(?)", v.(string)+"%") },
	"is_null":     func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f + " IS NULL") },
	"is_not_null": func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f + " IS NOT NULL") },
	"in":          func(q *gorm.DB, f string, v interface{}) *gorm.DB { return q.Where(f+" IN ?", v) },
}

func applyFilter(query *gorm.DB, table, field, operator string, value interface{}) *gorm.DB {
	fullField := `"` + table + `"."` + field + `"`

	applier, ok := filterOperators[operator]
	if ok {
		return applier(query, fullField, value)
	}
	return query
}

func (r *ReportRepository) applyProjectFilter(query *gorm.DB, projectID *int64, includeGlobal bool) *gorm.DB {
	if projectID == nil {
		return query
	}

	subquery := r.db.Table(`"Saved_Report_Project"`).Select("saved_report_id").Where("project_id = ?", *projectID)
	if !includeGlobal {
		return query.Where("id IN (?)", subquery)
	}

	noProjectSubquery := r.db.Table(`"Saved_Report_Project"`).Select("saved_report_id")
	return query.Where("id IN (?) OR id NOT IN (?)", subquery, noProjectSubquery)
}

func (r *ReportRepository) loadAuditUserInfo(result map[string]interface{}, tenantID int64, userID *int64) {
	if userID == nil {
		return
	}

	var user struct {
		ID        int64
		Email     string
		FirstName string
		LastName  string
	}
	err := r.db.Table(`"User"`).
		Select("id, email, first_name, last_name").
		Where("id = ? AND tenant_id = ?", *userID, tenantID).
		First(&user).Error
	if err != nil {
		return
	}

	result["user"] = map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.FirstName + " " + user.LastName,
	}
}
