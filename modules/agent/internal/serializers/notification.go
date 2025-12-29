package serializers

import (
	"fmt"
	"strings"
	"time"

	"agent/internal/models"
)

// EntityURLResolver provides methods to resolve entity URLs and project IDs
type EntityURLResolver interface {
	GetLeadType(entityID int64) string
	GetEntityProjectID(entityType string, entityID int64) *int64
}

// NotificationResponse represents a single notification
type NotificationResponse struct {
	ID        int64                    `json:"id"`
	Type      string                   `json:"type"`
	IsRead    bool                     `json:"is_read"`
	CreatedAt *string                  `json:"created_at"`
	Comment   *NotificationComment     `json:"comment"`
	Entity    *NotificationEntity      `json:"entity"`
}

// NotificationComment represents comment info in a notification
type NotificationComment struct {
	ID            *int64  `json:"id"`
	Content       string  `json:"content"`
	CreatedByName *string `json:"created_by_name"`
	CreatedAt     *string `json:"created_at"`
}

// NotificationEntity represents entity info in a notification
type NotificationEntity struct {
	Type        *string `json:"type"`
	ID          *int64  `json:"id"`
	DisplayName *string `json:"display_name"`
	URL         *string `json:"url"`
	ProjectID   *int64  `json:"project_id"`
}

// NotificationsResponse represents the list response
type NotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
}

var entityURLMap = map[string]string{
	"Lead":         "/leads",
	"Deal":         "/deals",
	"Task":         "/tasks",
	"Individual":   "/accounts/individuals",
	"Organization": "/accounts/organizations",
}

// NotificationToResponse converts a notification model to response
func NotificationToResponse(n *models.CommentNotification, resolver EntityURLResolver) NotificationResponse {
	resp := NotificationResponse{
		ID:     n.ID,
		Type:   n.NotificationType,
		IsRead: n.IsRead,
	}

	if !n.CreatedAt.IsZero() {
		t := n.CreatedAt.Format(time.RFC3339)
		resp.CreatedAt = &t
	}

	if n.Comment == nil {
		return resp
	}

	resp.Comment = buildNotificationComment(n.Comment)
	resp.Entity = buildNotificationEntity(n.Comment, resolver)

	return resp
}

func buildNotificationComment(c *models.Comment) *NotificationComment {
	comment := &NotificationComment{
		ID:      &c.ID,
		Content: truncateContent(c.Content, 100),
	}

	if !c.CreatedAt.IsZero() {
		t := c.CreatedAt.Format(time.RFC3339)
		comment.CreatedAt = &t
	}

	if c.Creator == nil {
		return comment
	}

	name := buildCreatorName(c.Creator)
	comment.CreatedByName = name

	return comment
}

func buildCreatorName(creator *models.User) *string {
	first := derefStr(creator.FirstName)
	last := derefStr(creator.LastName)
	name := strings.TrimSpace(first + " " + last)

	if name == "" {
		return &creator.Email
	}
	return &name
}

func buildNotificationEntity(c *models.Comment, resolver EntityURLResolver) *NotificationEntity {
	displayName := fmt.Sprintf("%s #%d", c.CommentableType, c.CommentableID)
	url := getEntityURL(c.CommentableType, c.CommentableID, &c.ID, resolver)
	projectID := resolver.GetEntityProjectID(c.CommentableType, c.CommentableID)

	return &NotificationEntity{
		Type:        &c.CommentableType,
		ID:          &c.CommentableID,
		DisplayName: &displayName,
		URL:         url,
		ProjectID:   projectID,
	}
}

func getEntityURL(entityType string, entityID int64, commentID *int64, resolver EntityURLResolver) *string {
	url := buildBaseEntityURL(entityType, entityID, resolver)
	url = appendCommentAnchor(url, commentID)
	return &url
}

func buildBaseEntityURL(entityType string, entityID int64, resolver EntityURLResolver) string {
	if entityType == "Lead" {
		leadType := resolver.GetLeadType(entityID)
		return fmt.Sprintf("/leads/%s/%d", leadType, entityID)
	}
	basePath := getEntityBasePath(entityType)
	return fmt.Sprintf("%s/%d", basePath, entityID)
}

func getEntityBasePath(entityType string) string {
	basePath, ok := entityURLMap[entityType]
	if ok {
		return basePath
	}
	return "/" + strings.ToLower(entityType) + "s"
}

func appendCommentAnchor(url string, commentID *int64) string {
	if commentID == nil {
		return url
	}
	return fmt.Sprintf("%s#comment-%d", url, *commentID)
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
