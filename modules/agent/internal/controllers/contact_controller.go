package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"agent/internal/middleware"
	"agent/internal/schemas"
	"agent/internal/serializers"
	"agent/internal/services"
)

// ContactController handles contact HTTP requests
type ContactController struct {
	contactService *services.ContactService
}

// NewContactController creates a new contact controller
func NewContactController(contactService *services.ContactService) *ContactController {
	return &ContactController{contactService: contactService}
}

// GetContacts handles GET /contacts
func (c *ContactController) GetContacts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	orderBy := r.URL.Query().Get("orderBy")
	order := r.URL.Query().Get("order")
	search := r.URL.Query().Get("search")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	result, err := c.contactService.GetContacts(page, limit, orderBy, order, search, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetContact handles GET /contacts/{id}
func (c *ContactController) GetContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact ID")
		return
	}

	result, err := c.contactService.GetContact(id)
	writeContactResponse(w, http.StatusOK, result, err)
}

// CreateContact handles POST /contacts
func (c *ContactController) CreateContact(w http.ResponseWriter, r *http.Request) {
	var input schemas.ContactCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.contactService.CreateContact(input, userID)
	writeContactResponse(w, http.StatusCreated, result, err)
}

// UpdateContact handles PUT /contacts/{id}
func (c *ContactController) UpdateContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact ID")
		return
	}

	var input schemas.ContactUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.contactService.UpdateContact(id, input, userID)
	writeContactResponse(w, http.StatusOK, result, err)
}

// DeleteContact handles DELETE /contacts/{id}
func (c *ContactController) DeleteContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact ID")
		return
	}

	err = c.contactService.DeleteContact(id)
	if errors.Is(err, services.ErrContactNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// SearchContacts handles GET /contacts/search
func (c *ContactController) SearchContacts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "search query is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	results, err := c.contactService.SearchContacts(query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func writeContactResponse(w http.ResponseWriter, successStatus int, result *serializers.ContactResponse, err error) {
	if errors.Is(err, services.ErrContactNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, services.ErrContactNameRequired) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, successStatus, result)
}
