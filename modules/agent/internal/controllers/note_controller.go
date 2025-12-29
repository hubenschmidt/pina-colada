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

type NoteController struct {
	noteService *services.NoteService
}

func NewNoteController(noteService *services.NoteService) *NoteController {
	return &NoteController{noteService: noteService}
}

func (c *NoteController) GetNotes(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")

	if entityType == "" || entityIDStr == "" {
		writeError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	notes, err := c.noteService.GetNotesByEntity(entityType, entityID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

func (c *NoteController) GetNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	note, err := c.noteService.GetNote(id)
	if errors.Is(err, services.ErrNoteNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, note)
}

type CreateNoteRequest struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	Content    string `json:"content"`
}

func (c *NoteController) CreateNote(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())

	input := services.NoteCreateInput{
		TenantID:   tenantID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Content:    req.Content,
		UserID:     userID,
	}

	note, err := c.noteService.CreateNote(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

type UpdateNoteRequest struct {
	Content string `json:"content"`
}

func (c *NoteController) UpdateNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	note, err := c.noteService.UpdateNote(id, req.Content, userID)
	if errors.Is(err, services.ErrNoteNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (c *NoteController) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	err = c.noteService.DeleteNote(id)
	if errors.Is(err, services.ErrNoteNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
