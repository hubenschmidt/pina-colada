package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	if err := runSeeders(ctx, conn, *seedersPath); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Seeding completed successfully")
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
		if strings.Contains(filepath.Base(file), "documents") {
			log.Printf("Skipping %s (handled by seed-documents)", filepath.Base(file))
			continue
		}

		if err := runSeederFile(ctx, conn, file); err != nil {
			return fmt.Errorf("failed to run %s: %w", filepath.Base(file), err)
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
