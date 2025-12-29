package services

import (
	"agent/internal/repositories"
	"agent/internal/serializers"
)

// NotificationService handles notification business logic
type NotificationService struct {
	notifRepo *repositories.NotificationRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(notifRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

// GetUnreadCount returns unread notification count for a user
func (s *NotificationService) GetUnreadCount(userID, tenantID int64) (int64, error) {
	return s.notifRepo.GetUnreadCount(userID, tenantID)
}

// GetNotifications returns notifications for a user
func (s *NotificationService) GetNotifications(userID, tenantID int64, limit int) (*serializers.NotificationsResponse, error) {
	notifications, err := s.notifRepo.GetNotifications(userID, tenantID, limit)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.notifRepo.GetUnreadCount(userID, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.NotificationResponse, len(notifications))
	for i := range notifications {
		items[i] = serializers.NotificationToResponse(&notifications[i], s.notifRepo)
	}

	return &serializers.NotificationsResponse{
		Notifications: items,
		UnreadCount:   unreadCount,
	}, nil
}

// MarkAsRead marks specific notifications as read
func (s *NotificationService) MarkAsRead(notificationIDs []int64, userID, tenantID int64) (int64, error) {
	return s.notifRepo.MarkAsRead(notificationIDs, userID, tenantID)
}

// MarkEntityAsRead marks all notifications for an entity as read
func (s *NotificationService) MarkEntityAsRead(userID, tenantID int64, entityType string, entityID int64) (int64, error) {
	return s.notifRepo.MarkEntityAsRead(userID, tenantID, entityType, entityID)
}
