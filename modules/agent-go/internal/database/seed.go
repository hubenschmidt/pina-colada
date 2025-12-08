package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SeederResult holds the result of a seeder run
type SeederResult struct {
	Applied []string
	Skipped []string
	Errors  []error
}

// EnsureSeedersTable creates the schema_seeders table if it doesn't exist
func EnsureSeedersTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_seeders (
			seeder_name TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		);
	`)
	return err
}

// GetAppliedSeeders returns the set of already-applied seeder names
func GetAppliedSeeders(db *sql.DB) (map[string]bool, error) {
	if err := EnsureSeedersTable(db); err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT seeder_name FROM schema_seeders")
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

// RecordSeeder records that a seeder has been applied
func RecordSeeder(db *sql.DB, name string) error {
	_, err := db.Exec(
		"INSERT INTO schema_seeders (seeder_name) VALUES ($1) ON CONFLICT DO NOTHING",
		name,
	)
	return err
}

// GetSeederFiles returns sorted list of .sql files in the seeders directory
func GetSeederFiles(seedersDir string) ([]string, error) {
	entries, err := os.ReadDir(seedersDir)
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

// RunSeeders applies all pending seeders
func RunSeeders(db *sql.DB, seedersDir string) (*SeederResult, error) {
	result := &SeederResult{}

	// Check RUN_SEEDERS environment variable
	runSeeders := strings.ToLower(os.Getenv("RUN_SEEDERS"))
	if runSeeders != "true" {
		fmt.Println("Skipping seeders (RUN_SEEDERS not set or not 'true')")
		return result, nil
	}

	// Ensure tracking table exists
	if err := EnsureSeedersTable(db); err != nil {
		return nil, fmt.Errorf("failed to ensure seeders table: %w", err)
	}

	// Get applied seeders
	applied, err := GetAppliedSeeders(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied seeders: %w", err)
	}

	// Get seeder files
	files, err := GetSeederFiles(seedersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get seeder files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No seeder files found")
		return result, nil
	}

	fmt.Printf("Found %d seeder file(s):\n", len(files))
	for _, f := range files {
		fmt.Printf("  - %s\n", f)
	}

	// Filter to unapplied seeders
	var pending []string
	for _, f := range files {
		if applied[f] {
			result.Skipped = append(result.Skipped, f)
		} else {
			pending = append(pending, f)
		}
	}

	if len(pending) == 0 {
		fmt.Println("All seeders already applied, nothing to run")
		return result, nil
	}

	fmt.Printf("\nFound %d seeder(s) to apply:\n", len(pending))
	for _, f := range pending {
		fmt.Printf("  - %s\n", f)
	}

	// Apply each pending seeder
	for _, filename := range pending {
		fmt.Printf("\nRunning seeder: %s\n", filename)

		content, err := os.ReadFile(filepath.Join(seedersDir, filename))
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to read %s: %w", filename, err))
			return result, result.Errors[len(result.Errors)-1]
		}

		res, err := db.Exec(string(content))
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to apply %s: %w", filename, err))
			return result, result.Errors[len(result.Errors)-1]
		}

		rowsAffected, _ := res.RowsAffected()

		if err := RecordSeeder(db, filename); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to record %s: %w", filename, err))
			return result, result.Errors[len(result.Errors)-1]
		}

		fmt.Printf("Seeder %s completed (%d rows affected)\n", filename, rowsAffected)
		result.Applied = append(result.Applied, filename)
	}

	fmt.Printf("\nAll %d seeder(s) completed successfully!\n", len(result.Applied))
	return result, nil
}

// GetSeedersDir returns the seeders directory path
func GetSeedersDir() string {
	// Docker path
	if _, err := os.Stat("/app/seeders"); err == nil {
		return "/app/seeders"
	}

	// Local development - relative to binary or working directory
	candidates := []string{
		"seeders",
		"../seeders",
		"../../seeders",
	}

	for _, dir := range candidates {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	return "seeders"
}
