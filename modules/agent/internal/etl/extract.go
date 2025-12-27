package etl

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

var TablesToExport = []string{
	"Tenant",
	"User",
	"Organization",
	"Lead",
	"Job",
	"Deal",
}

func ExportTables(ctx context.Context, conn *pgx.Conn, exportDir string) error {
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export dir: %w", err)
	}

	log.Printf("Exporting tables to %s/", exportDir)
	log.Println("--------------------------------------------------")

	var totalRows int
	for _, table := range TablesToExport {
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
			record[i] = fmt.Sprintf("%v", v)
			if v == nil {
				record[i] = ""
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
