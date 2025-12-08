package main

import (
	"fmt"
	"os"

	"github.com/pina-colada-co/agent-go/internal/database"
)

func printUsage() {
	fmt.Println(`Database Migration Tool

Usage:
  migrate <command>

Commands:
  up       Apply all pending migrations
  status   Check migration status
  seed     Run all pending seeders

Environment Variables:
  POSTGRES_HOST      Database host (default: localhost)
  POSTGRES_PORT      Database port (default: 5432)
  POSTGRES_USER      Database user
  POSTGRES_PASSWORD  Database password
  POSTGRES_DB        Database name
  RUN_SEEDERS        Set to 'true' to enable seeders`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runMigrations()
	case "status":
		checkStatus()
	case "seed":
		runSeeders()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runMigrations() {
	fmt.Println("=" + repeat("=", 59))
	fmt.Println("Database Migration Tool")
	fmt.Println("=" + repeat("=", 59))

	cfg := database.ConfigFromEnv()
	fmt.Printf("\nConnecting to %s:%s/%s...\n", cfg.Host, cfg.Port, cfg.Database)

	db, err := database.Connect(cfg, 3)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("Connected successfully!")

	migrationsDir := database.GetMigrationsDir()
	fmt.Printf("Migrations directory: %s\n", migrationsDir)

	result, err := database.ApplyMigrations(db, migrationsDir)
	if err != nil {
		fmt.Printf("\nMigration failed: %v\n", err)
		os.Exit(1)
	}

	if len(result.Applied) > 0 {
		fmt.Printf("\nApplied %d migration(s)\n", len(result.Applied))
	}
}

func checkStatus() {
	fmt.Println("=" + repeat("=", 59))
	fmt.Println("Migration Status Check")
	fmt.Println("=" + repeat("=", 59))

	cfg := database.ConfigFromEnv()
	fmt.Printf("\nConnecting to %s:%s/%s...\n", cfg.Host, cfg.Port, cfg.Database)

	db, err := database.Connect(cfg, 3)
	if err != nil {
		fmt.Printf("Could not connect to database: %v\n", err)
		fmt.Println("Cannot determine migration status")
		os.Exit(1)
	}
	defer db.Close()

	migrationsApplied, err := database.CheckMigrationStatus(db)
	if err != nil {
		fmt.Printf("Failed to check status: %v\n", err)
		os.Exit(1)
	}

	if migrationsApplied {
		fmt.Println("Database schema up to date, migrations already applied")
		return
	}

	migrationsDir := database.GetMigrationsDir()
	files, err := database.GetMigrationFiles(migrationsDir)
	if err != nil {
		fmt.Printf("Failed to read migrations: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No migration files found")
		return
	}

	applied, err := database.GetAppliedMigrations(db)
	if err != nil {
		fmt.Printf("Failed to get applied migrations: %v\n", err)
		os.Exit(1)
	}

	var pending []string
	for _, f := range files {
		if !applied[f] {
			pending = append(pending, f)
		}
	}

	if len(pending) == 0 {
		fmt.Println("All migrations have been applied")
	} else {
		fmt.Printf("%d migration(s) pending:\n", len(pending))
		for _, f := range pending {
			fmt.Printf("  - %s\n", f)
		}
	}
}

func runSeeders() {
	fmt.Println("=" + repeat("=", 59))
	fmt.Println("Database Seeder Tool")
	fmt.Println("=" + repeat("=", 59))

	cfg := database.ConfigFromEnv()
	fmt.Printf("\nConnecting to %s:%s/%s...\n", cfg.Host, cfg.Port, cfg.Database)

	db, err := database.Connect(cfg, 3)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("Connected successfully!")

	seedersDir := database.GetSeedersDir()
	fmt.Printf("Seeders directory: %s\n", seedersDir)

	result, err := database.RunSeeders(db, seedersDir)
	if err != nil {
		fmt.Printf("\nSeeder failed: %v\n", err)
		os.Exit(1)
	}

	if len(result.Applied) > 0 {
		fmt.Printf("\nApplied %d seeder(s)\n", len(result.Applied))
	}
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
