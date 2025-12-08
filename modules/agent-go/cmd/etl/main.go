package main

import (
	"fmt"
	"os"

	"github.com/pina-colada-co/agent-go/internal/database"
	"github.com/pina-colada-co/agent-go/internal/etl"
)

func printUsage() {
	fmt.Println(`ETL Tool - Extract, Transform, Load

Usage:
  etl <command> [options]

Commands:
  extract              Export tables to CSV files
  transform            Transform CSV schema (old -> new)
  load                 Import CSV files into database
  import-jobs <file>   Import jobs from CSV file
  seed-documents       Upload seed documents to S3

Options:
  --dry-run            Preview without making changes (import-jobs only)
  --no-duplicates      Skip duplicate checking (import-jobs only)
  --clear              Clear tables before loading (load only)
  --exports-dir <dir>  Directory for CSV exports (default: exports)
  --output-dir <dir>   Directory for transformed output (default: transformed)

Environment Variables:
  POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
  S3_BUCKET, AWS_REGION, S3_ENDPOINT (optional), AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY

Examples:
  etl extract
  etl transform --exports-dir exports --output-dir transformed
  etl load --clear
  etl import-jobs imports/jobs.csv --dry-run
  etl seed-documents`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "extract":
		runExtract()
	case "transform":
		runTransform()
	case "load":
		runLoad()
	case "import-jobs":
		runImportJobs()
	case "seed-documents":
		runSeedDocuments()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runExtract() {
	cfg := database.ConfigFromEnv()
	db, err := database.Connect(cfg, 3)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	exportsDir := getArg("--exports-dir", "exports")
	_, err = etl.ExportAll(db, exportsDir)
	if err != nil {
		fmt.Printf("Export failed: %v\n", err)
		os.Exit(1)
	}
}

func runTransform() {
	exportsDir := getArg("--exports-dir", "exports")
	outputDir := getArg("--output-dir", "transformed")

	if err := etl.Transform(exportsDir, outputDir); err != nil {
		fmt.Printf("Transform failed: %v\n", err)
		os.Exit(1)
	}
}

func runLoad() {
	cfg := database.ConfigFromEnv()
	db, err := database.Connect(cfg, 3)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	importDir := getArg("--output-dir", "transformed")
	clearFirst := hasFlag("--clear")

	_, err = etl.LoadAll(db, importDir, clearFirst)
	if err != nil {
		fmt.Printf("Load failed: %v\n", err)
		os.Exit(1)
	}
}

func runImportJobs() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: etl import-jobs <csv-file> [--dry-run] [--no-duplicates]")
		os.Exit(1)
	}

	csvFile := os.Args[2]
	dryRun := hasFlag("--dry-run")
	skipDuplicates := !hasFlag("--no-duplicates")

	cfg := database.ConfigFromEnv()
	db, err := database.Connect(cfg, 3)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println(repeat("=", 80))
	if dryRun {
		fmt.Print("[DRY RUN] ")
	}
	fmt.Printf("Importing jobs from: %s\n", csvFile)
	fmt.Println(repeat("=", 80))

	result, err := etl.ImportJobs(db, csvFile, dryRun, skipDuplicates)
	if err != nil {
		fmt.Printf("Import failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%sImport complete:\n", ternary(dryRun, "[DRY RUN] ", ""))
	fmt.Printf("  Imported: %d\n", result.Imported)
	fmt.Printf("  Skipped: %d\n", result.Skipped)
	if result.Duplicates > 0 {
		fmt.Printf("  Duplicates (skipped): %d\n", result.Duplicates)
	}
	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(result.Errors)-10)
				break
			}
			fmt.Printf("  - %s\n", err)
		}
	}
}

func runSeedDocuments() {
	cfg := database.ConfigFromEnv()
	db, err := database.Connect(cfg, 3)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	seedsDir := etl.GetSeedsDir()
	fmt.Printf("Starting seed document upload from %s...\n", seedsDir)

	result, err := etl.SeedDocuments(db, seedsDir)
	if err != nil {
		fmt.Printf("Seed documents failed: %v\n", err)
		os.Exit(1)
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
}

func getArg(flag, defaultVal string) string {
	for i, arg := range os.Args {
		if arg == flag && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return defaultVal
}

func hasFlag(flag string) bool {
	for _, arg := range os.Args {
		if arg == flag {
			return true
		}
	}
	return false
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
