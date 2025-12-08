package etl

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

// TablesToImport defines the import order (respects FK constraints)
var TablesToImport = []string{
	"Tenant",
	"Account",
	"Individual",
	"User",
	"Role",
	"User_Role",
	"Project",
	"Organization",
	"Deal",
	"Lead",
	"Lead_Project",
	"Job",
}

// TablesWithoutID are junction tables with composite keys
var TablesWithoutID = map[string]bool{
	"User_Role":    true,
	"Lead_Project": true,
}

// LoadResult holds the result of a load operation
type LoadResult struct {
	Table    string
	RowCount int
	Error    error
}

// LoadTable imports a single table from CSV file
func LoadTable(db *sql.DB, tableName, importDir string) (*LoadResult, error) {
	result := &LoadResult{Table: tableName}

	csvFile := filepath.Join(importDir, tableName+".csv")
	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		result.Error = fmt.Errorf("CSV file not found: %s", csvFile)
		return result, result.Error
	}

	// Read CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		result.Error = err
		return result, err
	}

	if len(records) < 2 {
		return result, nil // No data rows
	}

	headers := records[0]
	dataRows := records[1:]

	// Build INSERT statement
	quotedColumns := make([]string, len(headers))
	placeholders := make([]string, len(headers))
	for i, col := range headers {
		quotedColumns[i] = fmt.Sprintf(`"%s"`, col)
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	insertSQL := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s)`,
		tableName,
		strings.Join(quotedColumns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Insert each row
	for _, row := range dataRows {
		values := make([]interface{}, len(row))
		for i, val := range row {
			if val == "" {
				values[i] = nil
			} else {
				values[i] = val
			}
		}

		_, err := db.Exec(insertSQL, values...)
		if err != nil {
			result.Error = fmt.Errorf("insert failed: %w", err)
			return result, result.Error
		}
		result.RowCount++
	}

	return result, nil
}

// ResetSequence resets the sequence for a table to max(id) + 1
func ResetSequence(db *sql.DB, tableName string) error {
	if TablesWithoutID[tableName] {
		return nil // Skip tables without id column
	}

	sql := fmt.Sprintf(`
		SELECT setval(
			pg_get_serial_sequence('"%s"', 'id'),
			COALESCE((SELECT MAX(id) FROM "%s"), 1)
		)
	`, tableName, tableName)

	_, err := db.Exec(sql)
	return err
}

// ClearTable deletes all rows from a table
func ClearTable(db *sql.DB, tableName string) error {
	_, err := db.Exec(fmt.Sprintf(`DELETE FROM "%s"`, tableName))
	return err
}

// LoadAll imports all tables from CSV files
func LoadAll(db *sql.DB, importDir string, clearFirst bool) ([]*LoadResult, error) {
	fmt.Println(repeat("=", 60))
	fmt.Println("Production Data Import")
	fmt.Println(repeat("=", 60))
	fmt.Printf("Source: %s\n", importDir)
	fmt.Println()

	// Check if import directory exists
	if _, err := os.Stat(importDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("import directory not found: %s", importDir)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	fmt.Println("Connected to database successfully.")
	fmt.Println()

	// Clear existing data if requested
	if clearFirst {
		fmt.Println("Clearing existing data...")
		for i := len(TablesToImport) - 1; i >= 0; i-- {
			table := TablesToImport[i]
			if err := ClearTable(db, table); err != nil {
				fmt.Printf("  Warning clearing %s: %v\n", table, err)
			} else {
				fmt.Printf("  Cleared %s\n", table)
			}
		}
		fmt.Println()
	}

	// Import tables
	fmt.Println("Importing data...")
	var results []*LoadResult
	totalRows := 0

	for _, table := range TablesToImport {
		result, err := LoadTable(db, table, importDir)
		results = append(results, result)

		if err != nil {
			fmt.Printf("  %s: ERROR - %v\n", table, err)
			continue
		}

		fmt.Printf("  %s: %d rows - OK\n", table, result.RowCount)
		totalRows += result.RowCount

		// Reset sequence after import
		if result.RowCount > 0 {
			if err := ResetSequence(db, table); err != nil {
				fmt.Printf("    Warning: failed to reset sequence: %v\n", err)
			}
		}
	}

	fmt.Println()
	fmt.Printf("Total rows imported: %d\n", totalRows)
	fmt.Println(repeat("=", 60))
	fmt.Println("Import complete!")

	return results, nil
}
