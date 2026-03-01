package models

import (
	"time"

	pgvector "github.com/pgvector/pgvector-go"
)

type Embedding struct {
	ID         int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   int64            `gorm:"column:tenant_id;not null" json:"tenant_id"`
	SourceType string           `gorm:"column:source_type;not null" json:"source_type"`
	SourceID   int64            `gorm:"column:source_id;not null" json:"source_id"`
	ConfigID   *int64           `gorm:"column:config_id" json:"config_id"`
	ChunkIndex int              `gorm:"column:chunk_index;not null;default:0" json:"chunk_index"`
	ChunkText  string           `gorm:"column:chunk_text;not null" json:"chunk_text"`
	Embedding  pgvector.Vector  `gorm:"column:embedding;type:vector(3072);not null" json:"embedding"`
	CreatedAt  time.Time        `gorm:"autoCreateTime" json:"created_at"`
}

func (Embedding) TableName() string {
	return "Embedding"
}
