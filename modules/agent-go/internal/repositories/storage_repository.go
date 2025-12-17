package repositories

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// StorageRepository defines the interface for file storage operations
type StorageRepository interface {
	Upload(path string, data io.Reader, contentType string, size int64) error
	Download(path string) ([]byte, error)
	Delete(path string) error
	GetURL(path string) string
	Exists(path string) (bool, error)
}

// LocalStorageRepository implements filesystem storage for local development
type LocalStorageRepository struct {
	basePath string
}

// NewLocalStorageRepository creates a new local storage repository
func NewLocalStorageRepository(basePath string) *LocalStorageRepository {
	basePath = resolveBasePath(basePath)
	os.MkdirAll(basePath, 0755)
	log.Printf("LocalStorageRepository initialized with basePath: %s", basePath)
	return &LocalStorageRepository{basePath: basePath}
}

func resolveBasePath(basePath string) string {
	if basePath != "" {
		return basePath
	}
	if envPath := os.Getenv("STORAGE_PATH"); envPath != "" {
		return envPath
	}
	return "/app/storage/documents"
}

func (r *LocalStorageRepository) Upload(path string, data io.Reader, contentType string, size int64) error {
	fullPath := filepath.Join(r.basePath, path)
	log.Printf("LocalStorageRepository.Upload: basePath=%s path=%s fullPath=%s", r.basePath, path, fullPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Printf("LocalStorageRepository.Upload: MkdirAll failed: %v", err)
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		log.Printf("LocalStorageRepository.Upload: Create failed: %v", err)
		return err
	}
	defer file.Close()

	written, err := io.Copy(file, data)
	if err != nil {
		log.Printf("LocalStorageRepository.Upload: Copy failed: %v", err)
		return err
	}
	log.Printf("LocalStorageRepository.Upload: wrote %d bytes to %s", written, fullPath)
	return nil
}

func (r *LocalStorageRepository) Download(path string) ([]byte, error) {
	fullPath := filepath.Join(r.basePath, path)
	log.Printf("LocalStorageRepository.Download: %s", fullPath)
	return os.ReadFile(fullPath)
}

func (r *LocalStorageRepository) Delete(path string) error {
	fullPath := filepath.Join(r.basePath, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(fullPath)
}

func (r *LocalStorageRepository) GetURL(path string) string {
	return fmt.Sprintf("file://%s", filepath.Join(r.basePath, path))
}

func (r *LocalStorageRepository) Exists(path string) (bool, error) {
	_, err := os.Stat(filepath.Join(r.basePath, path))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// R2StorageRepository implements Cloudflare R2 storage (S3-compatible)
type R2StorageRepository struct {
	client *s3.S3
	bucket string
}

// NewR2StorageRepository creates a new R2 storage repository
func NewR2StorageRepository(accountID, accessKey, secretKey, bucket string) (*R2StorageRepository, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("auto"),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return &R2StorageRepository{
		client: s3.New(sess),
		bucket: bucket,
	}, nil
}

func (r *R2StorageRepository) Upload(path string, data io.Reader, contentType string, size int64) error {
	_, err := r.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(path),
		Body:          aws.ReadSeekCloser(data),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	return err
}

func (r *R2StorageRepository) Download(path string) ([]byte, error) {
	resp, err := r.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (r *R2StorageRepository) Delete(path string) error {
	_, err := r.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (r *R2StorageRepository) GetURL(path string) string {
	req, _ := r.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	url, _ := req.Presign(3600)
	return url
}

func (r *R2StorageRepository) Exists(path string) (bool, error) {
	_, err := r.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetStorageRepository returns the appropriate storage repository based on environment
func GetStorageRepository() StorageRepository {
	if os.Getenv("STORAGE_BACKEND") == "r2" {
		return mustInitR2Storage()
	}
	return NewLocalStorageRepository("")
}

func mustInitR2Storage() *R2StorageRepository {
	repo, err := NewR2StorageRepository(
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_SECRET_ACCESS_KEY"),
		os.Getenv("R2_BUCKET_NAME"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize R2 storage: %v", err))
	}
	return repo
}
