package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
)

func main() {
	seedersPath := flag.String("path", "/app/seeders", "path to seeders directory")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	if err := ensureSeedersTable(ctx, conn); err != nil {
		log.Fatalf("Failed to create seeders table: %v", err)
	}

	if err := runSeeders(ctx, conn, *seedersPath); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Seeding completed successfully")
}

// ensureSeedersTable creates the tracking table if it doesn't exist.
// This is a fallback - the table should be created by migration 076.
func ensureSeedersTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_seeders (
			seeder_name TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`)
	return err
}

func isSeederExecuted(ctx context.Context, conn *pgx.Conn, filename string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_seeders WHERE seeder_name = $1)`, filename).Scan(&exists)
	return exists, err
}

func recordSeeder(ctx context.Context, conn *pgx.Conn, filename string) error {
	_, err := conn.Exec(ctx, `INSERT INTO schema_seeders (seeder_name) VALUES ($1) ON CONFLICT DO NOTHING`, filename)
	return err
}

func runSeeders(ctx context.Context, conn *pgx.Conn, seedersPath string) error {
	files, err := filepath.Glob(filepath.Join(seedersPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to glob seeders: %w", err)
	}

	if len(files) == 0 {
		log.Println("No seeder files found")
		return nil
	}

	sort.Strings(files)

	for _, file := range files {
		filename := filepath.Base(file)

		executed, err := isSeederExecuted(ctx, conn, filename)
		if err != nil {
			return fmt.Errorf("failed to check seeder status: %w", err)
		}

		if executed {
			log.Printf("Skipping seeder (already executed): %s", filename)
			continue
		}

		if err := runSeederFile(ctx, conn, file); err != nil {
			return fmt.Errorf("failed to run %s: %w", filename, err)
		}

		if err := recordSeeder(ctx, conn, filename); err != nil {
			return fmt.Errorf("failed to record seeder %s: %w", filename, err)
		}
	}

	return nil
}

func runSeederFile(ctx context.Context, conn *pgx.Conn, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	log.Printf("Running seeder: %s", filepath.Base(file))

	_, err = conn.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	return nil
}
