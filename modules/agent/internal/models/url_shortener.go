package models

import "time"

type URLShortener struct {
	ID        int64     `gorm:"primaryKey"`
	Code      string    `gorm:"column:code;uniqueIndex"`
	FullURL   string    `gorm:"column:full_url"`
	CreatedAt time.Time `gorm:"column:created_at"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
}

func (URLShortener) TableName() string {
	return "URL_Shortener"
}
