package models

type Document struct {
	ID          int64  `gorm:"primaryKey" json:"id"` // FK to Asset.id
	StoragePath string `gorm:"not null" json:"storage_path"`
	FileSize    int64  `gorm:"not null" json:"file_size"`
}

func (Document) TableName() string {
	return "Document"
}
