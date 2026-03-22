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

	"agent/internal/filtering"
	"agent/internal/repositories"
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
	cleanup := flag.Bool("cleanup", false, "delete seed document rows from DB and storage")
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

	if *cleanup {
		cleanupSeedDocuments(ctx, conn)
		return
	}

	storage := repositories.GetStorageRepository()

	if err := uploadSeedDocuments(ctx, conn, storage, *seedersPath); err != nil {
		log.Fatalf("Seed documents failed: %v", err)
	}

	log.Println("Seed documents completed successfully")
}

func cleanupSeedDocuments(ctx context.Context, conn *pgx.Conn) {
	log.Println("Cleaning up seed document rows...")
	rows, err := conn.Query(ctx, `SELECT id FROM "Document" WHERE storage_path LIKE '%/seed/%'`)
	if err != nil {
		log.Printf("Cleanup query failed: %v", err)
		return
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Printf("Cleanup scan failed: %v", err)
			return
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		log.Println("No seed documents to clean up")
		return
	}

	queries := []string{
		`DELETE FROM "Entity_Tag" WHERE entity_type = 'Asset' AND entity_id = ANY($1)`,
		`DELETE FROM "Entity_Asset" WHERE asset_id = ANY($1)`,
		`DELETE FROM "Document" WHERE id = ANY($1)`,
		`DELETE FROM "Asset" WHERE id = ANY($1)`,
	}
	for _, q := range queries {
		if _, err := conn.Exec(ctx, q, ids); err != nil {
			log.Printf("Cleanup query failed: %v", err)
		}
	}
	log.Printf("Cleaned up %d seed documents", len(ids))
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
		uploaded, skipped, err := uploadTenantDocuments(ctx, conn, storage, seedersPath, tenantID)
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
	rows, err := conn.Query(ctx, `SELECT id FROM "Tenant" WHERE slug = 'pinacolada' ORDER BY id`)
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

func uploadTenantDocuments(ctx context.Context, conn *pgx.Conn, storage repositories.StorageRepository, seedersPath, tenantID string) (uploaded, skipped int, err error) {
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

		content, err := os.ReadFile(filePath)
		if err != nil {
			return uploaded, skipped, err
		}

		contentType := getContentType(filePath)
		if err := storage.Upload(storagePath, bytes.NewReader(content), contentType, int64(len(content))); err != nil {
			return uploaded, skipped, err
		}

		log.Printf("Uploaded: %s", storagePath)
		uploaded++

		text := filtering.ExtractDocumentContent(contentType, content)
		if text == "" {
			continue
		}
		_, err = conn.Exec(ctx,
			`UPDATE "Document" SET compressed = $1 WHERE storage_path = $2`,
			text, storagePath)
		if err != nil {
			log.Printf("Failed to update compressed for %s: %v", storagePath, err)
		}
	}

	return uploaded, skipped, nil
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
