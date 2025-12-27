package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/pina-colada-co/agent-go/internal/repositories"
)

var seedFiles = []string{
	"company_proposal.pdf",
	"meeting_notes.pdf",
	"contract_draft.pdf",
	"product_spec.pdf",
	"invoice_sample.pdf",
	"individual_resume.pdf",
	"william_hubenschmidt_resume.pdf",
	"william_hubenschmidt_summary.txt",
	"william_hubenschmidt_sample_answers.txt",
	"william_hubenschmidt_coverletter_1.pdf",
	"william_hubenschmidt_coverletter_2.pdf",
}

func main() {
	seedersPath := flag.String("path", "/app/seeders/documents", "path to seed documents directory")
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

	storage := repositories.GetStorageRepository()

	if err := uploadSeedDocuments(ctx, conn, storage, *seedersPath); err != nil {
		log.Fatalf("Seed documents failed: %v", err)
	}

	log.Println("Seed documents completed successfully")
}

func uploadSeedDocuments(ctx context.Context, conn *pgx.Conn, storage repositories.StorageRepository, seedersPath string) error {
	tenantIDs, err := getActiveTenants(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to get tenants: %w", err)
	}

	if len(tenantIDs) == 0 {
		log.Println("No active tenants found")
		return nil
	}

	var uploadedCount, skippedCount int

	for _, tenantID := range tenantIDs {
		uploaded, skipped, err := uploadTenantDocuments(storage, seedersPath, tenantID)
		if err != nil {
			return fmt.Errorf("failed for tenant %s: %w", tenantID, err)
		}
		uploadedCount += uploaded
		skippedCount += skipped
	}

	if uploadedCount > 0 {
		log.Printf("Seed file upload complete: %d uploaded, %d skipped", uploadedCount, skippedCount)
		return nil
	}

	if skippedCount > 0 {
		log.Printf("Seed files already exist (%d files), skipping upload", skippedCount)
	}

	return nil
}

func getActiveTenants(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	rows, err := conn.Query(ctx, `SELECT id FROM "Tenant" WHERE status = 'active' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenantIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		tenantIDs = append(tenantIDs, id)
	}

	return tenantIDs, rows.Err()
}

func uploadTenantDocuments(storage repositories.StorageRepository, seedersPath, tenantID string) (uploaded, skipped int, err error) {
	for _, filename := range seedFiles {
		filePath := filepath.Join(seedersPath, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("File not found: %s", filePath)
			continue
		}

		storagePath := fmt.Sprintf("%s/seed/%s", tenantID, filename)

		exists, err := storage.Exists(storagePath)
		if err != nil {
			return uploaded, skipped, err
		}
		if exists {
			skipped++
			continue
		}

		if err := uploadFile(storage, filePath, storagePath); err != nil {
			return uploaded, skipped, err
		}

		log.Printf("Uploaded: %s", storagePath)
		uploaded++
	}

	return uploaded, skipped, nil
}

func uploadFile(storage repositories.StorageRepository, filePath, storagePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentType := getContentType(filePath)
	return storage.Upload(storagePath, bytes.NewReader(content), contentType, int64(len(content)))
}

func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".pdf" {
		return "application/pdf"
	}

	if ext == ".txt" {
		return "text/plain"
	}

	return "application/octet-stream"
}
