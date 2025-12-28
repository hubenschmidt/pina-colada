package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

var TablesToLoad = []string{
	"Tenant",
	"Account",
	"User",
	"Individual",
	"Role",
	"User_Role",
	"User_Preferences",
	"Organization",
	"Project",
	"Deal",
	"Lead",
	"Lead_Project",
	"Job",
	"Comment",
	"Asset",
	"Document",
	"Entity_Asset",
}

func LoadTables(ctx context.Context, conn *pgx.Conn, inputDir string) error {
	log.Println("============================================================")
	log.Println("Loading transformed data")
	log.Println("============================================================")
	log.Printf("Source: %s", inputDir)
	log.Println()

	var totalRows int
	for _, table := range TablesToLoad {
		count, err := loadTable(ctx, conn, table, inputDir)
		if err != nil {
			log.Printf("  %s: ERROR - %v", table, err)
			continue
		}
		totalRows += count
	}

	log.Println()
	log.Println("============================================================")
	log.Printf("Load complete: %d total rows", totalRows)
	log.Println("============================================================")
	return nil
}

func loadTable(ctx context.Context, conn *pgx.Conn, table, inputDir string) (int, error) {
	csvPath := filepath.Join(inputDir, fmt.Sprintf("%s.csv", table))

	file, err := os.Open(csvPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("  %s: skipped (no file)", table)
			return 0, nil
		}
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		log.Printf("  %s: skipped (empty)", table)
		return 0, nil
	}

	headers := records[0]
	rows := records[1:]

	quotedHeaders := make([]string, len(headers))
	for i, h := range headers {
		quotedHeaders[i] = fmt.Sprintf(`"%s"`, h)
	}

	placeholders := make([]string, len(headers))
	for i := range headers {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s) ON CONFLICT DO NOTHING`,
		table,
		strings.Join(quotedHeaders, ", "),
		strings.Join(placeholders, ", "),
	)

	inserted := 0
	for _, row := range rows {
		args := make([]interface{}, len(row))
		for i, val := range row {
			if val == "" {
				args[i] = nil
			} else {
				args[i] = val
			}
		}

		_, err := conn.Exec(ctx, query, args...)
		if err != nil {
			log.Printf("  %s: row error - %v", table, err)
			continue
		}
		inserted++
	}

	log.Printf("  %s: %d rows loaded", table, inserted)
	return inserted, nil
}
