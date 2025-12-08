package etl

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SeedFiles defines the files to upload
var SeedFiles = []string{
	// Generic seed documents
	"company_proposal.pdf",
	"meeting_notes.pdf",
	"contract_draft.pdf",
	"product_spec.pdf",
	"invoice_sample.pdf",
	// Personal documents
	"william_hubenschmidt_resume.pdf",
	"william_hubenschmidt_summary.txt",
	"william_hubenschmidt_sample_answers.txt",
	"william_hubenschmidt_coverletter_1.pdf",
	"william_hubenschmidt_coverletter_2.pdf",
}

// ContentTypes maps file extensions to MIME types
var ContentTypes = map[string]string{
	".pdf": "application/pdf",
	".txt": "text/plain",
}

// SeedDocumentsResult holds the result of seeding documents
type SeedDocumentsResult struct {
	Uploaded int
	Skipped  int
	Errors   []string
}

// SeedDocuments uploads seed document files to S3 storage
func SeedDocuments(db *sql.DB, seedsDir string) (*SeedDocumentsResult, error) {
	result := &SeedDocumentsResult{}

	// Get S3 configuration from environment
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable not set")
	}

	// Create S3 client
	ctx := context.Background()

	var cfg aws.Config
	var err error

	if endpoint != "" {
		// Custom endpoint (e.g., MinIO, LocalStack)
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
			}, nil
		})

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// Get active tenants
	rows, err := db.Query(`SELECT id FROM "Tenant" WHERE status = 'active' ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenantIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		tenantIDs = append(tenantIDs, id)
	}

	if len(tenantIDs) == 0 {
		fmt.Println("No active tenants found")
		return result, nil
	}

	fmt.Printf("Found %d active tenant(s)\n", len(tenantIDs))

	// Upload seed files for each tenant
	for _, tenantID := range tenantIDs {
		for _, filename := range SeedFiles {
			localPath := filepath.Join(seedsDir, filename)
			if _, err := os.Stat(localPath); os.IsNotExist(err) {
				fmt.Printf("Warning: File not found: %s\n", localPath)
				continue
			}

			storagePath := fmt.Sprintf("%d/seed/%s", tenantID, filename)

			// Check if file already exists
			_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(storagePath),
			})
			if err == nil {
				result.Skipped++
				continue
			}

			// Read file content
			content, err := os.ReadFile(localPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", filename, err))
				continue
			}

			// Determine content type
			ext := strings.ToLower(filepath.Ext(filename))
			contentType := ContentTypes[ext]
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			// Upload to S3
			_, err = client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      aws.String(bucket),
				Key:         aws.String(storagePath),
				Body:        strings.NewReader(string(content)),
				ContentType: aws.String(contentType),
			})
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to upload %s: %v", storagePath, err))
				continue
			}

			fmt.Printf("Uploaded: %s\n", storagePath)
			result.Uploaded++
		}
	}

	if result.Uploaded > 0 {
		fmt.Printf("Seed file upload complete: %d uploaded, %d skipped\n", result.Uploaded, result.Skipped)
	} else if result.Skipped > 0 {
		fmt.Printf("Seed files already exist (%d files), skipping upload\n", result.Skipped)
	}

	return result, nil
}

// GetSeedsDir returns the seed documents directory path
func GetSeedsDir() string {
	// Docker path
	if _, err := os.Stat("/app/seeders/documents"); err == nil {
		return "/app/seeders/documents"
	}

	// Local development
	candidates := []string{
		"seeders/documents",
		"../seeders/documents",
		"../../seeders/documents",
	}

	for _, dir := range candidates {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	return "seeders/documents"
}
