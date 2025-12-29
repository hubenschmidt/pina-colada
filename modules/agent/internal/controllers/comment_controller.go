package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"agent/internal/middleware"
	"agent/internal/services"
)

type CommentController struct {
	commentService *services.CommentService
}

func NewCommentController(commentService *services.CommentService) *CommentController {
	return &CommentController{commentService: commentService}
}

func (c *CommentController) GetComments(w http.ResponseWriter, r *http.Request) {
	commentableType := r.URL.Query().Get("commentable_type")
	commentableIDStr := r.URL.Query().Get("commentable_id")

	if commentableType == "" || commentableIDStr == "" {
		writeError(w, http.StatusBadRequest, "commentable_type and commentable_id are required")
		return
	}

	commentableID, err := strconv.ParseInt(commentableIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid commentable_id")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	comments, err := c.commentService.GetCommentsByEntity(commentableType, commentableID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

func (c *CommentController) GetComment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment ID")
		return
	}

	comment, err := c.commentService.GetComment(id)
	if errors.Is(err, services.ErrCommentNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

type CreateCommentRequest struct {
	CommentableType string `json:"commentable_type"`
	CommentableID   int64  `json:"commentable_id"`
	Content         string `json:"content"`
	ParentCommentID *int64 `json:"parent_comment_id"`
}

func (c *CommentController) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())

	input := services.CommentCreateInput{
		TenantID:        tenantID,
		CommentableType: req.CommentableType,
		CommentableID:   req.CommentableID,
		Content:         req.Content,
		ParentCommentID: req.ParentCommentID,
		UserID:          userID,
	}

	comment, err := c.commentService.CreateComment(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

func (c *CommentController) UpdateComment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment ID")
		return
	}

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	comment, err := c.commentService.UpdateComment(id, req.Content, userID)
	if errors.Is(err, services.ErrCommentNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (c *CommentController) DeleteComment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment ID")
		return
	}

	err = c.commentService.DeleteComment(id)
	if errors.Is(err, services.ErrCommentNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
