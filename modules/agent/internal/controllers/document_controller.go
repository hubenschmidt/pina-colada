package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"agent/internal/lib"
	"agent/internal/middleware"
	"agent/internal/services"
)

type DocumentController struct {
	docService *services.DocumentService
}

func NewDocumentController(docService *services.DocumentService) *DocumentController {
	return &DocumentController{docService: docService}
}

func (c *DocumentController) GetDocuments(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	search := r.URL.Query().Get("search")

	entityID, err := lib.ParseOptionalInt64Param(r, "entity_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	var entityTypePtr *string
	if entityType != "" {
		entityTypePtr = &entityType
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	orderBy := r.URL.Query().Get("orderBy")
	order := r.URL.Query().Get("order")
	currentOnly := r.URL.Query().Get("current_only") == "true"

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	result, err := c.docService.GetDocuments(entityTypePtr, entityID, tenantID, search, page, limit, orderBy, order, currentOnly)
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
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"exists": false, "document": nil})
		return
	}

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	doc, err := c.docService.CheckFilenameExists(tenantID, filename)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if doc == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"exists": false, "document": nil})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"exists": true, "document": doc})
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
	entityIDPtr = parseOptionalInt64(entityIDStr)

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

// GetDocumentVersions handles GET /assets/documents/{id}/versions
func (c *DocumentController) GetDocumentVersions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	var tenantID int64 = 0
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = tid
	}

	versions, count, err := c.docService.GetDocumentVersions(id, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"versions":      versions,
		"version_count": count,
	})
}

// CreateDocumentVersion handles POST /assets/documents/{id}/versions
func (c *DocumentController) CreateDocumentVersion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	var tenantID int64 = 0
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = tid
	}

	userID, _ := middleware.GetUserID(r.Context())

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	doc, count, err := c.docService.CreateDocumentVersion(services.CreateDocumentVersionInput{
		DocumentID:  id,
		TenantID:    tenantID,
		UserID:      userID,
		Filename:    header.Filename,
		ContentType: contentType,
		Data:        file,
		Size:        header.Size,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"document":      doc,
		"version_count": count,
	})
}

// SetCurrentVersion handles PATCH /assets/documents/{id}/set-current
func (c *DocumentController) SetCurrentVersion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	var tenantID int64 = 0
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = tid
	}

	doc, count, err := c.docService.SetCurrentVersion(id, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document":      doc,
		"version_count": count,
	})
}

// DownloadDocument handles GET /assets/documents/{id}/download
func (c *DocumentController) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	var tenantID int64 = 0
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = tid
	}

	result, err := c.docService.DownloadDocument(id, tenantID)
	if errors.Is(err, services.ErrDocumentNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If we have a redirect URL, send redirect
	if result.RedirectURL != nil {
		http.Redirect(w, r, *result.RedirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Otherwise, send the content directly
	w.Header().Set("Content-Type", result.Document.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+result.Document.Filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(result.Content)
}

// UpdateDocument handles PUT /assets/documents/{id}
func (c *DocumentController) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	var input services.UpdateDocumentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var tenantID int64 = 0
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = tid
	}
	userID, _ := middleware.GetUserID(r.Context())

	doc, err := c.docService.UpdateDocument(id, tenantID, userID, input)
	if errors.Is(err, services.ErrDocumentNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

// DeleteDocument handles DELETE /assets/documents/{id}
func (c *DocumentController) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document ID")
		return
	}

	var tenantID int64 = 0
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = tid
	}

	err = c.docService.DeleteDocument(id, tenantID)
	if errors.Is(err, services.ErrDocumentNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
