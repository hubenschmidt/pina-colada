package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	extractCmd := flag.NewFlagSet("extract", flag.ExitOnError)
	extractDir := extractCmd.String("dir", "./cmd/etl/exports", "output directory for CSV files")

	transformCmd := flag.NewFlagSet("transform", flag.ExitOnError)
	transformInput := transformCmd.String("input", "./cmd/etl/exports", "input directory with source CSVs")
	transformOutput := transformCmd.String("output", "./cmd/etl/transformed", "output directory for transformed CSVs")

	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)
	loadDir := loadCmd.String("dir", "./cmd/etl/transformed", "directory with transformed CSV files")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	dbURL := ""
	if dbURL == "" {
		log.Fatal("ETL_DATABASE_URL environment variable is required")
	}

	switch os.Args[1] {
	case "extract":
		extractCmd.Parse(os.Args[2:])
		runExtract(dbURL, *extractDir)

	case "transform":
		transformCmd.Parse(os.Args[2:])
		runTransform(*transformInput, *transformOutput)

	case "load":
		loadCmd.Parse(os.Args[2:])
		runLoad(dbURL, *loadDir)

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
	fmt.Println("  load          Load transformed CSVs into database")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  etl extract")
	fmt.Println("  etl transform")
	fmt.Println("  etl load")
}

func runExtract(dbURL, exportDir string) {
	ctx := context.Background()
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	if err := ExportTables(ctx, conn, exportDir); err != nil {
		log.Fatalf("Extract failed: %v", err)
	}
}

func runTransform(inputDir, outputDir string) {
	if err := Transform(inputDir, outputDir); err != nil {
		log.Fatalf("Transform failed: %v", err)
	}
}

func runLoad(dbURL, inputDir string) {
	ctx := context.Background()
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	if err := LoadTables(ctx, conn, inputDir); err != nil {
		log.Fatalf("Load failed: %v", err)
	}
}
