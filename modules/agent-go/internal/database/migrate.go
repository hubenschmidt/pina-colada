package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// MigrationResult holds the result of a migration run
type MigrationResult struct {
	Applied []string
	Skipped []string
	Errors  []error
}

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// ConfigFromEnv creates a Config from environment variables
func ConfigFromEnv() Config {
	sslMode := os.Getenv("POSTGRES_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	return Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
		SSLMode:  sslMode,
	}
}

// DSN returns the PostgreSQL connection string
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// Connect establishes a database connection with retries
func Connect(cfg Config, maxRetries int) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		db, err = sql.Open("postgres", cfg.DSN())
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err == nil {
			return db, nil
		}

		db.Close()
		if attempt < maxRetries-1 {
			fmt.Printf("Connection attempt %d failed, retrying...\n", attempt+1)
			time.Sleep(2 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

// EnsureMigrationsTable creates the schema_migrations table if it doesn't exist
func EnsureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			migration_name TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		);
	`)
	return err
}

// GetAppliedMigrations returns the set of already-applied migration names
func GetAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	if err := EnsureMigrationsTable(db); err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT migration_name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, rows.Err()
}

// RecordMigration records that a migration has been applied
func RecordMigration(db *sql.DB, name string) error {
	_, err := db.Exec(
		"INSERT INTO schema_migrations (migration_name) VALUES ($1) ON CONFLICT DO NOTHING",
		name,
	)
	return err
}

// GetMigrationFiles returns sorted list of .sql files in the migrations directory
func GetMigrationFiles(migrationsDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}

// ApplyMigrations applies all pending migrations
func ApplyMigrations(db *sql.DB, migrationsDir string) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Get applied migrations
	applied, err := GetAppliedMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get migration files
	files, err := GetMigrationFiles(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No migration files found")
		return result, nil
	}

	// Filter to unapplied migrations
	var pending []string
	for _, f := range files {
		if applied[f] {
			result.Skipped = append(result.Skipped, f)
		} else {
			pending = append(pending, f)
		}
	}

	if len(pending) == 0 {
		fmt.Println("Database schema up to date, no migrations needed")
		return result, nil
	}

	fmt.Printf("Found %d migration(s) to apply:\n", len(pending))
	for _, f := range pending {
		fmt.Printf("  - %s\n", f)
	}

	// Apply each pending migration
	for _, filename := range pending {
		fmt.Printf("\nApplying: %s\n", filename)

		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to read %s: %w", filename, err))
			return result, result.Errors[len(result.Errors)-1]
		}

		_, err = db.Exec(string(content))
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to apply %s: %w", filename, err))
			return result, result.Errors[len(result.Errors)-1]
		}

		if err := RecordMigration(db, filename); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to record %s: %w", filename, err))
			return result, result.Errors[len(result.Errors)-1]
		}

		fmt.Printf("Successfully applied %s\n", filename)
		result.Applied = append(result.Applied, filename)
	}

	fmt.Printf("\nAll %d migration(s) applied successfully!\n", len(result.Applied))
	return result, nil
}

// CheckMigrationStatus checks if migrations have been applied
func CheckMigrationStatus(db *sql.DB) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'Job'
		)
	`).Scan(&exists)

	return exists, err
}

// GetMigrationsDir returns the migrations directory path
func GetMigrationsDir() string {
	// Docker path
	if _, err := os.Stat("/app/migrations"); err == nil {
		return "/app/migrations"
	}

	// Local development - relative to binary or working directory
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
	}

	for _, dir := range candidates {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	return "migrations"
}
