package repositories

import (
	"errors"

	apperrors "agent/internal/errors"
	"agent/internal/models"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) FindByEntity(commentableType string, commentableID int64, tenantID *int64) ([]models.Comment, error) {
	var comments []models.Comment
	query := r.db.Preload("Creator").Where("commentable_type = ? AND commentable_id = ?", commentableType, commentableID)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	err := query.Order("created_at DESC").Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) FindByID(id int64) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("Creator").First(&comment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

type CommentCreateInput struct {
	TenantID        int64
	CommentableType string
	CommentableID   int64
	Content         string
	ParentCommentID *int64
	UserID          int64
}

func (r *CommentRepository) Create(input CommentCreateInput) (int64, error) {
	comment := &models.Comment{
		TenantID:        input.TenantID,
		CommentableType: input.CommentableType,
		CommentableID:   input.CommentableID,
		Content:         input.Content,
		ParentCommentID: input.ParentCommentID,
		CreatedBy:       input.UserID,
		UpdatedBy:       input.UserID,
	}
	err := r.db.Create(comment).Error
	return comment.ID, err
}

func (r *CommentRepository) Update(id int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *CommentRepository) Delete(id int64) error {
	return r.db.Delete(&models.Comment{}, id).Error
}
