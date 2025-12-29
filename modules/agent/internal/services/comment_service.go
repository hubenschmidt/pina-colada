package services

import (
	"errors"

	"agent/internal/repositories"
	"agent/internal/serializers"
)

var ErrCommentNotFound = errors.New("comment not found")

type CommentService struct {
	commentRepo *repositories.CommentRepository
}

func NewCommentService(commentRepo *repositories.CommentRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo}
}

func (s *CommentService) GetCommentsByEntity(commentableType string, commentableID int64, tenantID *int64) ([]serializers.CommentResponse, error) {
	comments, err := s.commentRepo.FindByEntity(commentableType, commentableID, tenantID)
	if err != nil {
		return nil, err
	}
	return serializers.CommentsToResponse(comments), nil
}

func (s *CommentService) GetComment(id int64) (*serializers.CommentResponse, error) {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	resp := serializers.CommentToResponse(comment)
	return &resp, nil
}

type CommentCreateInput struct {
	TenantID        int64
	CommentableType string
	CommentableID   int64
	Content         string
	ParentCommentID *int64
	UserID          int64
}

func (s *CommentService) CreateComment(input CommentCreateInput) (*serializers.CommentResponse, error) {
	repoInput := repositories.CommentCreateInput{
		TenantID:        input.TenantID,
		CommentableType: input.CommentableType,
		CommentableID:   input.CommentableID,
		Content:         input.Content,
		ParentCommentID: input.ParentCommentID,
		UserID:          input.UserID,
	}
	id, err := s.commentRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}
	return s.GetComment(id)
}

func (s *CommentService) UpdateComment(id int64, content string, userID int64) (*serializers.CommentResponse, error) {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	updates := map[string]interface{}{
		"content":    content,
		"updated_by": userID,
	}
	if err := s.commentRepo.Update(id, updates); err != nil {
		return nil, err
	}
	return s.GetComment(id)
}

func (s *CommentService) DeleteComment(id int64) error {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	return s.commentRepo.Delete(id)
}
