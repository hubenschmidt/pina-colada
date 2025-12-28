package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
)

func ExportTables(ctx context.Context, conn *pgx.Conn, exportDir string) error {
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export dir: %w", err)
	}

	tables, err := getNonEmptyTables(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	log.Printf("Exporting %d non-empty tables to %s/", len(tables), exportDir)
	log.Println("--------------------------------------------------")

	var totalRows int
	for _, table := range tables {
		count, err := exportTable(ctx, conn, table, exportDir)
		if err != nil {
			log.Printf("  %s: ERROR - %v", table, err)
			continue
		}
		totalRows += count
	}

	log.Println("--------------------------------------------------")
	log.Printf("Export complete: %d total rows", totalRows)
	return nil
}

func getNonEmptyTables(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	query := `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename
	`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var allTables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, err
		}
		allTables = append(allTables, name)
	}
	rows.Close()

	var tables []string
	for _, name := range allTables {
		var hasData bool
		existsQuery := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM "%s" LIMIT 1)`, name)
		if err := conn.QueryRow(ctx, existsQuery).Scan(&hasData); err != nil {
			log.Printf("  Skipping %s: %v", name, err)
			continue
		}
		if hasData {
			tables = append(tables, name)
		}
	}
	return tables, nil
}

func exportTable(ctx context.Context, conn *pgx.Conn, table, exportDir string) (int, error) {
	outputFile := filepath.Join(exportDir, fmt.Sprintf("%s.csv", table))

	rows, err := conn.Query(ctx, fmt.Sprintf(`SELECT * FROM "%s"`, table))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	file, err := os.Create(outputFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	fieldDescs := rows.FieldDescriptions()
	headers := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		headers[i] = string(fd.Name)
	}
	if err := writer.Write(headers); err != nil {
		return 0, err
	}

	rowCount := 0
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return rowCount, err
		}

		record := make([]string, len(values))
		for i, v := range values {
			switch val := v.(type) {
			case nil:
				record[i] = ""
			case time.Time:
				record[i] = val.Format(time.RFC3339)
			default:
				record[i] = fmt.Sprintf("%v", v)
			}
		}

		if err := writer.Write(record); err != nil {
			return rowCount, err
		}
		rowCount++
	}

	if rowCount > 0 {
		log.Printf("  %s: %d rows -> %s", table, rowCount, filepath.Base(outputFile))
	} else {
		log.Printf("  %s: 0 rows (skipped)", table)
	}

	return rowCount, nil
}
