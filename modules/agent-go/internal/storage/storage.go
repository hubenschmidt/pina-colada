package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Backend defines the storage interface
type Backend interface {
	Upload(path string, data io.Reader, contentType string, size int64) error
	Download(path string) ([]byte, error)
	Delete(path string) error
	GetURL(path string) string
	Exists(path string) (bool, error)
}

// LocalStorage implements filesystem storage for local development
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage backend
func NewLocalStorage(basePath string) *LocalStorage {
	if basePath == "" {
		basePath = os.Getenv("STORAGE_PATH")
		if basePath == "" {
			basePath = "/app/storage/documents"
		}
	}
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Upload(path string, data io.Reader, contentType string, size int64) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	return err
}

func (s *LocalStorage) Download(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.basePath, path))
}

func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(fullPath)
}

func (s *LocalStorage) GetURL(path string) string {
	return fmt.Sprintf("file://%s", filepath.Join(s.basePath, path))
}

func (s *LocalStorage) Exists(path string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.basePath, path))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// R2Storage implements Cloudflare R2 storage (S3-compatible)
type R2Storage struct {
	client *s3.S3
	bucket string
}

// NewR2Storage creates a new R2 storage backend
func NewR2Storage(accountID, accessKey, secretKey, bucket string) (*R2Storage, error) {
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

	return &R2Storage{
		client: s3.New(sess),
		bucket: bucket,
	}, nil
}

func (s *R2Storage) Upload(path string, data io.Reader, contentType string, size int64) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(path),
		Body:          aws.ReadSeekCloser(data),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	return err
}

func (s *R2Storage) Download(path string) ([]byte, error) {
	resp, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *R2Storage) Delete(path string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (s *R2Storage) GetURL(path string) string {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	url, _ := req.Presign(3600)
	return url
}

func (s *R2Storage) Exists(path string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetStorage returns the appropriate storage backend based on environment
func GetStorage() Backend {
	backend := os.Getenv("STORAGE_BACKEND")

	if backend == "r2" {
		storage, err := NewR2Storage(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_SECRET_ACCESS_KEY"),
			os.Getenv("R2_BUCKET_NAME"),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize R2 storage: %v", err))
		}
		return storage
	}

	return NewLocalStorage("")
}
