package models

import "time"

type CommentNotification struct {
	ID               int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID         int64      `gorm:"not null;index" json:"tenant_id"`
	UserID           int64      `gorm:"not null;index" json:"user_id"`
	CommentID        int64      `gorm:"not null;index" json:"comment_id"`
	NotificationType string     `gorm:"size:20;not null" json:"notification_type"` // direct_reply, thread_activity
	IsRead           bool       `gorm:"default:false" json:"is_read"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	ReadAt           *time.Time `json:"read_at"`
}

func (CommentNotification) TableName() string {
	return "Comment_Notification"
}
