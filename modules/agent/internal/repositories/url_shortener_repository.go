package repositories

import (
	"time"

	"gorm.io/gorm"

	"agent/internal/models"
)

type URLShortenerRepository struct {
	db *gorm.DB
}

func NewURLShortenerRepository(db *gorm.DB) *URLShortenerRepository {
	return &URLShortenerRepository{db: db}
}

// Store saves a URL shortcode mapping
func (r *URLShortenerRepository) Store(code, fullURL string, ttl time.Duration) error {
	var existing models.URLShortener
	err := r.db.Where("code = ?", code).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		entry := models.URLShortener{
			Code:      code,
			FullURL:   fullURL,
			ExpiresAt: time.Now().Add(ttl),
		}
		return r.db.Create(&entry).Error
	}
	if err != nil {
		return err
	}

	return r.db.Model(&existing).Updates(map[string]interface{}{
		"full_url":   fullURL,
		"expires_at": time.Now().Add(ttl),
	}).Error
}

// Lookup retrieves the full URL for a code
func (r *URLShortenerRepository) Lookup(code string) (string, error) {
	var entry models.URLShortener
	err := r.db.Where("code = ? AND expires_at > ?", code, time.Now()).First(&entry).Error

	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return entry.FullURL, nil
}

// DeleteExpired removes expired entries
func (r *URLShortenerRepository) DeleteExpired() (int64, error) {
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&models.URLShortener{})
	return result.RowsAffected, result.Error
}
