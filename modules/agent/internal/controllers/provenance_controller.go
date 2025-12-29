package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"agent/internal/middleware"
	"agent/internal/schemas"
	"agent/internal/services"
)

// ProvenanceController handles provenance HTTP requests
type ProvenanceController struct {
	provenanceService *services.ProvenanceService
}

// NewProvenanceController creates a new provenance controller
func NewProvenanceController(provenanceService *services.ProvenanceService) *ProvenanceController {
	return &ProvenanceController{provenanceService: provenanceService}
}

// GetProvenance handles GET /provenance/{entity_type}/{entity_id}
func (c *ProvenanceController) GetProvenance(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityIDStr := chi.URLParam(r, "entityID")

	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	var fieldName *string
	if fn := r.URL.Query().Get("field_name"); fn != "" {
		fieldName = &fn
	}

	records, err := c.provenanceService.GetProvenance(entityType, entityID, fieldName)
	if errors.Is(err, services.ErrInvalidEntityType) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, records)
}

// CreateProvenance handles POST /provenance
func (c *ProvenanceController) CreateProvenance(w http.ResponseWriter, r *http.Request) {
	var input schemas.ProvenanceCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.EntityType == "" || input.FieldName == "" || input.Source == "" {
		writeError(w, http.StatusBadRequest, "entity_type, field_name, and source are required")
		return
	}

	var userID *int64
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		userID = &uid
	}

	record, err := c.provenanceService.CreateProvenance(input, userID)
	if errors.Is(err, services.ErrInvalidEntityType) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, record)
}
