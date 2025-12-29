package repositories

import (
	"strings"
	"time"

	"agent/internal/models"
	"gorm.io/gorm"
)

// NotificationRepository handles notification data access
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(userID, tenantID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.CommentNotification{}).
		Where("user_id = ? AND tenant_id = ? AND is_read = ?", userID, tenantID, false).
		Count(&count).Error
	return count, err
}

// GetNotifications returns notifications for a user with comment details
func (r *NotificationRepository) GetNotifications(userID, tenantID int64, limit int) ([]models.CommentNotification, error) {
	var notifications []models.CommentNotification
	err := r.db.Model(&models.CommentNotification{}).
		Preload("Comment").
		Preload("Comment.Creator").
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// MarkAsRead marks specific notifications as read
func (r *NotificationRepository) MarkAsRead(notificationIDs []int64, userID, tenantID int64) (int64, error) {
	now := time.Now()
	result := r.db.Model(&models.CommentNotification{}).
		Where("id IN ? AND user_id = ? AND tenant_id = ? AND is_read = ?", notificationIDs, userID, tenantID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})
	return result.RowsAffected, result.Error
}

// MarkEntityAsRead marks all notifications for an entity as read
func (r *NotificationRepository) MarkEntityAsRead(userID, tenantID int64, entityType string, entityID int64) (int64, error) {
	commentIDs := r.db.Model(&models.Comment{}).
		Select("id").
		Where("commentable_type = ? AND commentable_id = ?", entityType, entityID)

	now := time.Now()
	result := r.db.Model(&models.CommentNotification{}).
		Where("user_id = ? AND tenant_id = ? AND comment_id IN (?) AND is_read = ?", userID, tenantID, commentIDs, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})
	return result.RowsAffected, result.Error
}

// GetLeadType returns the lead type for URL building
func (r *NotificationRepository) GetLeadType(entityID int64) string {
	var leadType string
	r.db.Model(&models.Lead{}).
		Select("type").
		Where("id = ?", entityID).
		Scan(&leadType)

	if leadType == "" {
		return "leads"
	}

	if leadType == "Opportunity" {
		return "opportunities"
	}
	return pluralizeType(leadType)
}

func pluralizeType(t string) string {
	if t == "" {
		return "leads"
	}
	return strings.ToLower(t) + "s"
}

// GetEntityProjectID returns the project_id for an entity
func (r *NotificationRepository) GetEntityProjectID(entityType string, entityID int64) *int64 {
	if entityType == "" || entityID == 0 {
		return nil
	}

	handlers := map[string]func(int64) *int64{
		"Deal":         r.getDealProjectID,
		"Lead":         r.getLeadProjectID,
		"Task":         r.getTaskProjectID,
		"Individual":   r.getAccountEntityProjectID,
		"Organization": r.getAccountEntityProjectID,
	}

	handler, ok := handlers[entityType]
	if !ok {
		return nil
	}
	return handler(entityID)
}

func (r *NotificationRepository) getDealProjectID(entityID int64) *int64 {
	var projectID int64
	r.db.Model(&models.Deal{}).Select("project_id").Where("id = ?", entityID).Scan(&projectID)
	return positiveOrNil(projectID)
}

func (r *NotificationRepository) getLeadProjectID(entityID int64) *int64 {
	var projectID int64
	r.db.Table("Lead_Project").Select("project_id").Where("lead_id = ?", entityID).Limit(1).Scan(&projectID)
	return positiveOrNil(projectID)
}

func (r *NotificationRepository) getTaskProjectID(entityID int64) *int64 {
	var task struct {
		TaskableType *string
		TaskableID   *int64
	}
	r.db.Model(&models.Task{}).Select("taskable_type, taskable_id").Where("id = ?", entityID).Scan(&task)
	if task.TaskableType == nil || task.TaskableID == nil {
		return nil
	}
	return r.GetEntityProjectID(*task.TaskableType, *task.TaskableID)
}

func (r *NotificationRepository) getAccountEntityProjectID(entityID int64) *int64 {
	var accountID int64
	r.db.Model(&models.Individual{}).Select("account_id").Where("id = ?", entityID).Scan(&accountID)
	if accountID == 0 {
		r.db.Model(&models.Organization{}).Select("account_id").Where("id = ?", entityID).Scan(&accountID)
	}
	if accountID == 0 {
		return nil
	}

	var projectID int64
	r.db.Table("Account_Project").Select("project_id").Where("account_id = ?", accountID).Limit(1).Scan(&projectID)
	return positiveOrNil(projectID)
}

func positiveOrNil(val int64) *int64 {
	if val <= 0 {
		return nil
	}
	return &val
}
