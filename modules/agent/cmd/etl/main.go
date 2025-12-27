package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pina-colada-co/agent-go/internal/etl"
)

func main() {
	extractCmd := flag.NewFlagSet("extract", flag.ExitOnError)
	extractDir := extractCmd.String("dir", "./exports", "output directory for CSV files")

	transformCmd := flag.NewFlagSet("transform", flag.ExitOnError)
	transformInput := transformCmd.String("input", "./exports", "input directory with source CSVs")
	transformOutput := transformCmd.String("output", "./transformed", "output directory for transformed CSVs")

	importCmd := flag.NewFlagSet("import-jobs", flag.ExitOnError)
	importFile := importCmd.String("file", "", "path to CSV file (required)")
	importDryRun := importCmd.Bool("dry-run", false, "preview import without saving")
	importNoDupes := importCmd.Bool("no-skip-duplicates", false, "allow duplicate imports")
	importTenantID := importCmd.Int64("tenant-id", 1, "tenant ID for imported jobs")
	importDealID := importCmd.Int64("deal-id", 1, "deal ID for imported jobs")
	importUserID := importCmd.Int64("user-id", 1, "user ID for created_by/updated_by")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	switch os.Args[1] {
	case "extract":
		extractCmd.Parse(os.Args[2:])
		runExtract(dbURL, *extractDir)

	case "transform":
		transformCmd.Parse(os.Args[2:])
		runTransform(*transformInput, *transformOutput)

	case "import-jobs":
		importCmd.Parse(os.Args[2:])
		if *importFile == "" {
			log.Fatal("--file is required")
		}
		runImportJobs(dbURL, *importFile, *importDryRun, !*importNoDupes, *importTenantID, *importDealID, *importUserID)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ETL Tool - Extract, Transform, Load utilities")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  etl <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  extract       Export database tables to CSV files")
	fmt.Println("  transform     Transform CSVs from old schema to new schema")
	fmt.Println("  import-jobs   Import jobs from CSV file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  etl extract --dir ./exports")
	fmt.Println("  etl transform --input ./exports --output ./transformed")
	fmt.Println("  etl import-jobs --file ./imports/jobs.csv --dry-run")
}

func runExtract(dbURL, exportDir string) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	if err := etl.ExportTables(ctx, conn, exportDir); err != nil {
		log.Fatalf("Extract failed: %v", err)
	}
}

func runTransform(inputDir, outputDir string) {
	if err := etl.Transform(inputDir, outputDir); err != nil {
		log.Fatalf("Transform failed: %v", err)
	}
}

func runImportJobs(dbURL, csvPath string, dryRun, skipDuplicates bool, tenantID, dealID, userID int64) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	prefix := ""
	if dryRun {
		prefix = "[DRY RUN] "
	}
	log.Printf("%sImporting jobs from: %s", prefix, csvPath)
	log.Println("================================================================================")

	stats, err := etl.ImportJobs(db, csvPath, dryRun, skipDuplicates, tenantID, dealID, userID)
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	log.Printf("\n%sImport complete:", prefix)
	log.Printf("  Imported: %d", stats.Imported)
	log.Printf("  Skipped: %d", stats.Skipped)
	if stats.Duplicates > 0 {
		log.Printf("  Duplicates (skipped): %d", stats.Duplicates)
	}

	if len(stats.Errors) > 0 {
		log.Printf("\nErrors (%d):", len(stats.Errors))
		limit := 10
		if len(stats.Errors) < limit {
			limit = len(stats.Errors)
		}
		for _, e := range stats.Errors[:limit] {
			log.Printf("  - %s", e)
		}
		if len(stats.Errors) > 10 {
			log.Printf("  ... and %d more", len(stats.Errors)-10)
		}
	}
}
