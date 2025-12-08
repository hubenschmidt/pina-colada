package etl

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

// TablesToExport defines the tables to export in dependency order
var TablesToExport = []string{
	"Tenant",
	"User",
	"Organization",
	"Lead",
	"Job",
	"Deal",
}

// ExportResult holds the result of an export operation
type ExportResult struct {
	Table    string
	RowCount int
	Error    error
}

// ExportTable exports a single table to CSV file
func ExportTable(db *sql.DB, tableName, outputDir string) (*ExportResult, error) {
	result := &ExportResult{Table: tableName}

	// Query all rows from table
	query := fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	rows, err := db.Query(query)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		result.Error = err
		return result, err
	}

	// Create output file
	outputPath := filepath.Join(outputDir, tableName+".csv")
	file, err := os.Create(outputPath)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(columns); err != nil {
		result.Error = err
		return result, err
	}

	// Prepare value holders
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Write rows
	rowCount := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			result.Error = err
			return result, err
		}

		record := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}

		if err := writer.Write(record); err != nil {
			result.Error = err
			return result, err
		}
		rowCount++
	}

	if err := rows.Err(); err != nil {
		result.Error = err
		return result, err
	}

	result.RowCount = rowCount
	return result, nil
}

// ExportAll exports all tables to CSV files
func ExportAll(db *sql.DB, outputDir string) ([]*ExportResult, error) {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Exporting tables to %s/\n", outputDir)
	fmt.Println(repeat("-", 50))

	var results []*ExportResult
	totalRows := 0

	for _, table := range TablesToExport {
		result, err := ExportTable(db, table, outputDir)
		results = append(results, result)

		if err != nil {
			fmt.Printf("  %s: ERROR - %v\n", table, err)
			continue
		}

		if result.RowCount > 0 {
			fmt.Printf("  %s: %d rows -> %s.csv\n", table, result.RowCount, table)
			totalRows += result.RowCount
		} else {
			fmt.Printf("  %s: 0 rows (skipped)\n", table)
		}
	}

	fmt.Println(repeat("-", 50))
	fmt.Printf("Export complete: %d total rows\n", totalRows)

	return results, nil
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
