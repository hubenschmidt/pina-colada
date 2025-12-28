package models

import "time"

type Comment struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        int64     `gorm:"not null;index" json:"tenant_id"`
	CommentableType string    `gorm:"size:50;not null" json:"commentable_type"`
	CommentableID   int64     `gorm:"not null" json:"commentable_id"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	ParentCommentID *int64    `gorm:"index" json:"parent_comment_id"`
	CreatedBy       int64     `gorm:"not null" json:"created_by"`
	UpdatedBy       int64     `gorm:"not null" json:"updated_by"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Comment) TableName() string {
	return "Comment"
}
