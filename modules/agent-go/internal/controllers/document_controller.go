package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/services"
)

type DocumentController struct {
	docService *services.DocumentService
}

func NewDocumentController(docService *services.DocumentService) *DocumentController {
	return &DocumentController{docService: docService}
}

func (c *DocumentController) GetDocuments(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")
	search := r.URL.Query().Get("search")

	var entityID *int64
	if entityIDStr != "" {
		id, err := strconv.ParseInt(entityIDStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid entity_id")
			return
		}
		entityID = &id
	}

	var entityTypePtr *string
	if entityType != "" {
		entityTypePtr = &entityType
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	orderBy := r.URL.Query().Get("orderBy")
	order := r.URL.Query().Get("order")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	result, err := c.docService.GetDocuments(entityTypePtr, entityID, tenantID, search, page, limit, orderBy, order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetDocument handles GET /assets/documents/{id}
func (c *DocumentController) GetDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	doc, err := c.docService.GetDocumentByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if doc == nil {
		writeError(w, http.StatusNotFound, "document not found")
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

// CheckFilename handles GET /assets/documents/check-filename
func (c *DocumentController) CheckFilename(w http.ResponseWriter, r *http.Request) {
	// For now, always return not exists to allow upload
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"exists":   false,
		"document": nil,
	})
}

// UploadDocument handles POST /assets/documents
func (c *DocumentController) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	// Get the file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Get optional form fields
	description := r.FormValue("description")
	entityType := r.FormValue("entity_type")
	entityIDStr := r.FormValue("entity_id")

	var descPtr *string
	if description != "" {
		descPtr = &description
	}

	var entityTypePtr *string
	var entityIDPtr *int64
	if entityType != "" {
		entityTypePtr = &entityType
	}
	if entityIDStr != "" {
		id, err := strconv.ParseInt(entityIDStr, 10, 64)
		if err == nil {
			entityIDPtr = &id
		}
	}

	// Get tenant and user ID from context
	var tenantID int64 = 0
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = tid
	}

	userID, _ := middleware.GetUserID(r.Context())

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload document
	result, err := c.docService.UploadDocument(services.UploadDocumentInput{
		TenantID:    tenantID,
		UserID:      userID,
		Filename:    header.Filename,
		ContentType: contentType,
		Data:        file,
		Size:        header.Size,
		Description: descPtr,
		EntityType:  entityTypePtr,
		EntityID:    entityIDPtr,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// LinkDocument handles POST /assets/documents/{id}/link
func (c *DocumentController) LinkDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	var input services.EntityLinkInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := c.docService.LinkDocument(id, input); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// UnlinkDocument handles DELETE /assets/documents/{id}/link
func (c *DocumentController) UnlinkDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

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

	if err := c.docService.UnlinkDocument(id, entityType, entityID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
