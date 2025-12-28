package models

import "gorm.io/datatypes"

// DocumentSummary represents an AI-generated summary of a document
type DocumentSummary struct {
	Text        string   `json:"text"`
	Keywords    []string `json:"keywords,omitempty"`
	GeneratedAt string   `json:"generated_at,omitempty"`
}

type Document struct {
	ID          int64          `gorm:"primaryKey" json:"id"` // FK to Asset.id
	StoragePath string         `gorm:"not null" json:"storage_path"`
	FileSize    int64          `gorm:"not null" json:"file_size"`
	Summary     datatypes.JSON `gorm:"column:summary" json:"summary,omitempty"`
	Compressed  *string        `gorm:"column:compressed" json:"compressed,omitempty"`
}

func (Document) TableName() string {
	return "Document"
}
