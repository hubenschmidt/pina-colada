package etl

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// JobImportResult holds the result of a job import
type JobImportResult struct {
	Imported   int
	Skipped    int
	Duplicates int
	Errors     []string
}

// JobRecord represents a job to import
type JobRecord struct {
	Company     string
	JobTitle    string
	JobURL      string
	Notes       string
	ResumeDate  *time.Time
	Source      string
	DateApplied *time.Time
}

// ImportJobs imports jobs from a CSV file
func ImportJobs(db *sql.DB, csvPath string, dryRun bool, skipDuplicates bool) (*JobImportResult, error) {
	result := &JobImportResult{}

	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return result, nil // No data
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Load existing jobs for duplicate check
	var existingJobs []JobRecord
	if skipDuplicates && !dryRun {
		fmt.Println("Loading existing jobs to check for duplicates...")
		existingJobs, err = loadExistingJobs(db)
		if err != nil {
			fmt.Printf("Warning: Could not load existing jobs: %v\n", err)
			fmt.Println("Continuing without duplicate check...")
		} else {
			fmt.Printf("  Found %d existing jobs\n", len(existingJobs))
		}
	}

	// Process rows
	for rowNum, row := range records[1:] {
		job, skip := parseJobRow(row, headerMap, rowNum+2)
		if skip {
			result.Skipped++
			continue
		}

		if skipDuplicates && isDuplicateJob(job, existingJobs) {
			fmt.Printf("⊘ Row %d: %s - %s (duplicate, skipping)\n", rowNum+2, job.Company, job.JobTitle)
			result.Duplicates++
			continue
		}

		if !dryRun {
			if err := insertJob(db, job); err != nil {
				errMsg := fmt.Sprintf("Row %d: %v", rowNum+2, err)
				result.Errors = append(result.Errors, errMsg)
				fmt.Printf("✗ %s\n", errMsg)
				result.Skipped++
				continue
			}
		}

		existingJobs = append(existingJobs, *job)
		fmt.Printf("✓ Row %d: %s - %s\n", rowNum+2, job.Company, job.JobTitle)
		result.Imported++
	}

	return result, nil
}

func parseJobRow(row []string, headerMap map[string]int, rowNum int) (*JobRecord, bool) {
	getField := func(keys ...string) string {
		for _, key := range keys {
			if idx, ok := headerMap[strings.ToLower(key)]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
		}
		return ""
	}

	description := getField("description", "company")
	jobTitle := getField("job title", "job_title", "title")
	dateStr := getField("date applied", "date", "date_applied", "application_date")

	// Try to split description if no job title
	if description != "" && jobTitle == "" {
		if company, title := trySplitDescription(description); company != "" {
			description = company
			jobTitle = title
		}
	}

	// Skip if missing required fields
	if description == "" || jobTitle == "" {
		return nil, true
	}

	// Skip non-applied entries
	dateLower := strings.ToLower(dateStr)
	skipIndicators := []string{"did not apply", "to apply", "jen to apply", "need to finish"}
	for _, skip := range skipIndicators {
		if dateLower == skip {
			return nil, true
		}
	}

	job := &JobRecord{
		Company:  description,
		JobTitle: jobTitle,
		JobURL:   getField("link", "url", "job_url", "job_link"),
		Notes:    getField("note", "notes"),
		Source:   coalesce(getField("job board", "job_board", "source"), "manual"),
	}

	// Parse dates
	if dateStr != "" {
		if parsed := parseDate(dateStr); parsed != nil {
			job.DateApplied = parsed
		}
	}

	resumeStr := getField("resume")
	if resumeStr != "" {
		if parsed := parseDate(resumeStr); parsed != nil {
			job.ResumeDate = parsed
		} else if job.Notes != "" {
			job.Notes = job.Notes + " | Resume: " + resumeStr
		} else {
			job.Notes = "Resume: " + resumeStr
		}
	}

	return job, false
}

func trySplitDescription(description string) (string, string) {
	// Try em-dash or dash with space
	re := regexp.MustCompile(`\s*[–-]\s+`)
	parts := re.Split(description, 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	// Try em-dash only
	if strings.Contains(description, "–") {
		parts := strings.SplitN(description, "–", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	// Try dash with spaces
	if strings.Contains(description, " - ") {
		parts := strings.SplitN(description, " - ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	return "", ""
}

func parseDate(dateStr string) *time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return nil
	}

	// Skip non-date values
	skipValues := []string{"did not apply", "to apply", "jen to apply", "need to finish"}
	for _, skip := range skipValues {
		if strings.ToLower(dateStr) == skip {
			return nil
		}
	}

	formats := []string{
		"1/2/2006",  // 11/1/2024
		"1/2",       // 11/1 (assume current year)
		"1/2/06",    // 11/1/24
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// If no year in format, use current year
			if format == "1/2" {
				t = t.AddDate(time.Now().Year()-t.Year(), 0, 0)
			}
			return &t
		}
	}

	return nil
}

func isDuplicateJob(job *JobRecord, existing []JobRecord) bool {
	for _, e := range existing {
		if strings.ToLower(e.Company) != strings.ToLower(job.Company) {
			continue
		}
		if strings.ToLower(e.JobTitle) != strings.ToLower(job.JobTitle) {
			continue
		}
		// Check date match
		if job.DateApplied != nil && e.DateApplied != nil {
			if job.DateApplied.Format("2006-01-02") == e.DateApplied.Format("2006-01-02") {
				return true
			}
		} else if job.DateApplied == nil && e.DateApplied == nil {
			return true
		}
	}
	return false
}

func loadExistingJobs(db *sql.DB) ([]JobRecord, error) {
	rows, err := db.Query(`
		SELECT l.title, j.job_title, l.created_at
		FROM "Lead" l
		JOIN "Job" j ON j.id = l.id
		WHERE l.type = 'Job'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobRecord
	for rows.Next() {
		var company, jobTitle string
		var createdAt time.Time
		if err := rows.Scan(&company, &jobTitle, &createdAt); err != nil {
			continue
		}
		jobs = append(jobs, JobRecord{
			Company:     company,
			JobTitle:    jobTitle,
			DateApplied: &createdAt,
		})
	}

	return jobs, rows.Err()
}

func insertJob(db *sql.DB, job *JobRecord) error {
	// This is a simplified version - in production you'd need to:
	// 1. Create or find the Deal
	// 2. Create the Lead
	// 3. Create the Job
	// For now, just insert into Job table directly

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create Deal
	var dealID int64
	err = tx.QueryRow(`
		INSERT INTO "Deal" (tenant_id, name, description, created_at, updated_at, created_by, updated_by)
		VALUES (1, $1, '', NOW(), NOW(), 1, 1)
		RETURNING id
	`, job.Company+" - "+job.JobTitle).Scan(&dealID)
	if err != nil {
		return fmt.Errorf("failed to create deal: %w", err)
	}

	// Create Lead
	var leadID int64
	err = tx.QueryRow(`
		INSERT INTO "Lead" (tenant_id, deal_id, type, title, source, created_at, updated_at, created_by, updated_by)
		VALUES (1, $1, 'Job', $2, $3, NOW(), NOW(), 1, 1)
		RETURNING id
	`, dealID, job.Company, job.Source).Scan(&leadID)
	if err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}

	// Create Job
	_, err = tx.Exec(`
		INSERT INTO "Job" (id, organization_id, job_title, job_url, notes, resume_date, created_at, updated_at)
		VALUES ($1, 1, $2, $3, $4, $5, NOW(), NOW())
	`, leadID, job.JobTitle, job.JobURL, job.Notes, job.ResumeDate)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return tx.Commit()
}
