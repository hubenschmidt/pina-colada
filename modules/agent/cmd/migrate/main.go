package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	migrationsPath := flag.String("path", "/app/migrations", "path to migrations directory")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"up"}
	}

	command := args[0]

	m, err := migrate.New(
		fmt.Sprintf("file://%s", *migrationsPath),
		dbURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	// Log each migration step
	m.Log = &migrateLogger{}

	if err := runCommand(m, command); err != nil {
		log.Fatalf("Migration %s failed: %v", command, err)
	}

	log.Printf("Migration %s completed successfully", command)
}

func runCommand(m *migrate.Migrate, command string) error {
	if command == "up" {
		return handleMigrateError(m.Up())
	}

	if command == "down" {
		return handleMigrateError(m.Down())
	}

	if command == "status" {
		return printStatus(m)
	}

	return fmt.Errorf("unknown command: %s (valid: up, down, status)", command)
}

func handleMigrateError(err error) error {
	if err == migrate.ErrNoChange {
		log.Println("No migrations to apply")
		return nil
	}
	return err
}

func printStatus(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		log.Println("No migrations applied yet")
		return nil
	}
	if err != nil {
		return err
	}

	log.Printf("Current version: %d, dirty: %v", version, dirty)
	return nil
}

// migrateLogger implements migrate.Logger
type migrateLogger struct{}

func (l *migrateLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *migrateLogger) Verbose() bool {
	return true
}
