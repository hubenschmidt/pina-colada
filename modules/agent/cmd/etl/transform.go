package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
)

// Tables that need created_by/updated_by audit fields added
var tablesNeedingAuditFields = []string{
	"Account",
	"Individual",
	"Lead",
	"Deal",
	"Organization",
}

// Columns to remove from specific tables
var columnsToRemove = map[string][]string{
	"Lead": {"title", "description"},
}

func Transform(inputDir, outputDir string) error {
	log.Println("============================================================")
	log.Println("CSV Schema Transformation (Passthrough + Additions)")
	log.Println("============================================================")
	log.Printf("Source: %s", inputDir)
	log.Printf("Output: %s", outputDir)
	log.Println()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Get default user ID for audit fields
	defaultUserID := getDefaultUserID(inputDir)
	log.Printf("Default user_id for audit fields: %s", defaultUserID)
	log.Println()

	// Find all CSV files in input directory
	files, err := filepath.Glob(filepath.Join(inputDir, "*.csv"))
	if err != nil {
		return fmt.Errorf("failed to list CSV files: %w", err)
	}

	for _, file := range files {
		tableName := fileToTableName(file)
		if err := transformTable(file, outputDir, tableName, defaultUserID); err != nil {
			log.Printf("  %s: ERROR - %v", tableName, err)
			continue
		}
	}

	log.Println()
	log.Println("============================================================")
	log.Println("Transformation complete!")
	log.Printf("Output files in: %s", outputDir)
	log.Println("============================================================")
	return nil
}

func getDefaultUserID(inputDir string) string {
	users := readCSV(inputDir, "User.csv")
	if len(users) > 0 {
		return users[0]["id"]
	}
	return "1"
}

func fileToTableName(path string) string {
	base := filepath.Base(path)
	return base[:len(base)-4] // remove .csv
}

func transformTable(inputFile, outputDir, tableName, defaultUserID string) error {
	records, headers, err := readCSVWithHeaders(inputFile)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		log.Printf("  %s: skipped (empty)", tableName)
		return nil
	}

	// Determine transformations needed
	needsAuditFields := slices.Contains(tablesNeedingAuditFields, tableName)
	removeColumns := columnsToRemove[tableName]

	// Build new headers
	newHeaders := make([]string, 0, len(headers)+2)
	for _, h := range headers {
		if !slices.Contains(removeColumns, h) {
			newHeaders = append(newHeaders, h)
		}
	}
	if needsAuditFields {
		if !slices.Contains(newHeaders, "created_by") {
			newHeaders = append(newHeaders, "created_by")
		}
		if !slices.Contains(newHeaders, "updated_by") {
			newHeaders = append(newHeaders, "updated_by")
		}
	}

	// Transform records
	var newRecords []map[string]string
	for _, record := range records {
		newRecord := make(map[string]string)
		for _, h := range newHeaders {
			if h == "created_by" || h == "updated_by" {
				if val, ok := record[h]; ok && val != "" {
					newRecord[h] = val
				} else {
					newRecord[h] = defaultUserID
				}
			} else {
				newRecord[h] = record[h]
			}
		}
		newRecords = append(newRecords, newRecord)
	}

	// Write output
	outputFile := filepath.Join(outputDir, tableName+".csv")
	if err := writeCSVFile(outputFile, newRecords, newHeaders); err != nil {
		return err
	}

	changes := ""
	if needsAuditFields {
		changes += " +audit"
	}
	if len(removeColumns) > 0 {
		changes += fmt.Sprintf(" -%v", removeColumns)
	}
	if changes == "" {
		changes = " (passthrough)"
	}
	log.Printf("  %s: %d rows%s", tableName, len(newRecords), changes)
	return nil
}

func readCSVWithHeaders(path string) ([]map[string]string, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	allRecords, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(allRecords) < 1 {
		return nil, nil, nil
	}

	headers := allRecords[0]
	var result []map[string]string
	for _, row := range allRecords[1:] {
		m := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		result = append(result, m)
	}
	return result, headers, nil
}

func writeCSVFile(path string, data []map[string]string, headers []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range data {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = row[h]
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// Keep for backwards compatibility
func readCSV(dir, filename string) []map[string]string {
	path := filepath.Join(dir, filename)
	records, _, _ := readCSVWithHeaders(path)
	return records
}
