package etl

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/pina-colada-co/agent-go/internal/models"
)

type ImportStats struct {
	Imported   int
	Skipped    int
	Duplicates int
	Errors     []string
}

type jobImportData struct {
	Company     string
	JobTitle    string
	JobURL      string
	Description string
	ResumeDate  *time.Time
	Source      string
	AppliedDate *time.Time
}

func ImportJobs(db *gorm.DB, csvPath string, dryRun, skipDuplicates bool, tenantID, dealID, userID int64) (*ImportStats, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return &ImportStats{}, nil
	}

	headers := normalizeHeaders(records[0])
	var existingJobs []models.Job
	if skipDuplicates && !dryRun {
		log.Println("Loading existing jobs to check for duplicates...")
		db.Find(&existingJobs)
		log.Printf("  Found %d existing jobs", len(existingJobs))
	}

	stats := &ImportStats{}

	for i, record := range records[1:] {
		rowNum := i + 2
		row := mapRow(headers, record)

		if shouldSkipRow(row) {
			stats.Skipped++
			continue
		}

		data := parseJobRow(row)
		if data == nil {
			stats.Skipped++
			continue
		}

		if skipDuplicates && isDuplicate(data, existingJobs) {
			log.Printf("⊘ Row %d: %s - %s (duplicate, skipping)", rowNum, data.Company, data.JobTitle)
			stats.Duplicates++
			continue
		}

		if !dryRun {
			if err := createJobRecord(db, data, tenantID, dealID, userID); err != nil {
				errMsg := fmt.Sprintf("Row %d: %v", rowNum, err)
				log.Printf("✗ %s", errMsg)
				stats.Errors = append(stats.Errors, errMsg)
				stats.Skipped++
				continue
			}
		}

		log.Printf("✓ Row %d: %s - %s", rowNum, data.Company, data.JobTitle)
		stats.Imported++
	}

	return stats, nil
}

func normalizeHeaders(headers []string) map[string]int {
	result := make(map[string]int)
	for i, h := range headers {
		result[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return result
}

func mapRow(headers map[string]int, record []string) map[string]string {
	row := make(map[string]string)
	for key, idx := range headers {
		if idx < len(record) {
			row[key] = strings.TrimSpace(record[idx])
		}
	}
	return row
}

func getField(row map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := row[strings.ToLower(k)]; ok && v != "" {
			return v
		}
	}
	return ""
}

func shouldSkipRow(row map[string]string) bool {
	desc := getField(row, "description", "company")
	title := getField(row, "job title", "job_title", "title")
	dateApplied := strings.ToLower(getField(row, "date applied", "date", "date_applied"))

	if desc == "" || (title == "" && !strings.Contains(desc, " - ")) {
		return true
	}

	skipIndicators := []string{"did not apply", "to apply", "jen to apply", "need to finish"}
	for _, s := range skipIndicators {
		if dateApplied == s {
			return true
		}
	}
	return false
}

func parseJobRow(row map[string]string) *jobImportData {
	company := getField(row, "description", "company")
	jobTitle := getField(row, "job title", "job_title", "title")
	dateStr := getField(row, "date applied", "date", "date_applied")
	jobURL := getField(row, "link", "url", "job_url")
	notes := getField(row, "note", "notes")
	resumeStr := getField(row, "resume", "resume")
	source := getField(row, "job board", "job_board", "source")
	if source == "" {
		source = "manual"
	}

	if jobTitle == "" && strings.Contains(company, " - ") {
		parts := splitDescription(company)
		if parts != nil {
			company, jobTitle = parts[0], parts[1]
		}
	}

	if company == "" || jobTitle == "" {
		return nil
	}

	var resumeDate *time.Time
	if resumeStr != "" {
		if d := parseDate(resumeStr); d != nil {
			resumeDate = d
		} else if notes != "" {
			notes = fmt.Sprintf("%s | Resume: %s", notes, resumeStr)
		} else {
			notes = fmt.Sprintf("Resume: %s", resumeStr)
		}
	}

	var appliedDate *time.Time
	if dateStr != "" {
		appliedDate = parseDate(dateStr)
	}

	return &jobImportData{
		Company:     company,
		JobTitle:    jobTitle,
		JobURL:      jobURL,
		Description: notes,
		ResumeDate:  resumeDate,
		Source:      source,
		AppliedDate: appliedDate,
	}
}

func splitDescription(desc string) []string {
	patterns := []string{`\s*[–-]\s+`, ` – `, ` - `}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		parts := re.Split(desc, 2)
		if len(parts) == 2 {
			return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
		}
	}
	return nil
}

func parseDate(dateStr string) *time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return nil
	}

	skip := []string{"did not apply", "to apply", "jen to apply", "need to finish"}
	for _, s := range skip {
		if strings.ToLower(dateStr) == s {
			return nil
		}
	}

	formats := []string{"1/2/2006", "1/2", "1/2/06"}
	for _, fmt := range formats {
		if t, err := time.Parse(fmt, dateStr); err == nil {
			if fmt == "1/2" {
				t = t.AddDate(time.Now().Year(), 0, 0)
			}
			return &t
		}
	}
	return nil
}

func isDuplicate(data *jobImportData, existing []models.Job) bool {
	for _, e := range existing {
		if strings.EqualFold(e.JobTitle, data.JobTitle) {
			return true
		}
	}
	return false
}

func createJobRecord(db *gorm.DB, data *jobImportData, tenantID, dealID, userID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Create Account for the company
		account := models.Account{
			TenantID:  &tenantID,
			Name:      data.Company,
			CreatedBy: userID,
			UpdatedBy: userID,
		}
		if err := tx.Create(&account).Error; err != nil {
			return fmt.Errorf("failed to create account: %w", err)
		}

		// Create Lead
		lead := models.Lead{
			TenantID:  &tenantID,
			AccountID: &account.ID,
			DealID:    dealID,
			Type:      "Job",
			Source:    strPtr(data.Source),
			CreatedBy: userID,
			UpdatedBy: userID,
		}
		if data.AppliedDate != nil {
			lead.CreatedAt = *data.AppliedDate
		}
		if err := tx.Create(&lead).Error; err != nil {
			return fmt.Errorf("failed to create lead: %w", err)
		}

		// Create Job
		job := models.Job{
			ID:       lead.ID,
			JobTitle: data.JobTitle,
			JobURL:   strPtr(data.JobURL),
		}
		if data.Description != "" {
			job.Description = &data.Description
		}
		if data.ResumeDate != nil {
			job.ResumeDate = data.ResumeDate
		}
		if err := tx.Create(&job).Error; err != nil {
			return fmt.Errorf("failed to create job: %w", err)
		}

		return nil
	})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
