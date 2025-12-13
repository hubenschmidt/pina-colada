package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/pina-colada-co/agent-go/internal/lib"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// IndividualController handles individual HTTP requests
type IndividualController struct {
	indService *services.IndividualService
}

// NewIndividualController creates a new individual controller
func NewIndividualController(indService *services.IndividualService) *IndividualController {
	return &IndividualController{indService: indService}
}

// GetIndividuals handles GET /individuals
func (c *IndividualController) GetIndividuals(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	orderBy := r.URL.Query().Get("orderBy")
	order := r.URL.Query().Get("order")
	search := r.URL.Query().Get("search")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	result, err := c.indService.GetIndividuals(page, limit, orderBy, order, search, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetIndividual handles GET /individuals/{id}
func (c *IndividualController) GetIndividual(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	result, err := c.indService.GetIndividual(id)
	if errors.Is(err, services.ErrIndividualNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SearchIndividuals handles GET /individuals/search
func (c *IndividualController) SearchIndividuals(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	limit := lib.ParseIntQueryParam(r, "limit", 10)

	results, err := c.indService.SearchIndividuals(query, tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// AddContact handles POST /individuals/{id}/contacts
func (c *IndividualController) AddContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	var input services.ContactCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.indService.AddContactToIndividual(id, input, userID)
	if errors.Is(err, services.ErrIndividualNotFound) || errors.Is(err, services.ErrIndividualNoAccount) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// UpdateIndividual handles PUT /individuals/{id}
func (c *IndividualController) UpdateIndividual(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	var input services.IndividualUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.indService.UpdateIndividual(id, input, userID)
	if errors.Is(err, services.ErrIndividualNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteIndividual handles DELETE /individuals/{id}
func (c *IndividualController) DeleteIndividual(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	err = c.indService.DeleteIndividual(id)
	if errors.Is(err, services.ErrIndividualNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// CreateIndividual handles POST /individuals
func (c *IndividualController) CreateIndividual(w http.ResponseWriter, r *http.Request) {
	var input schemas.IndividualCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}
	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.indService.CreateIndividual(input, tenantID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetContacts handles GET /individuals/{id}/contacts
func (c *IndividualController) GetContacts(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	contacts, err := c.indService.GetContacts(id)
	if errors.Is(err, services.ErrIndividualNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

// UpdateContact handles PUT /individuals/{id}/contacts/{contactId}
func (c *IndividualController) UpdateContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	contactIDStr := chi.URLParam(r, "contactId")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact ID")
		return
	}

	var input schemas.IndContactUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.indService.UpdateContact(id, contactID, input, userID)
	if errors.Is(err, services.ErrIndividualNotFound) || errors.Is(err, services.ErrIndividualContactNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteContact handles DELETE /individuals/{id}/contacts/{contactId}
func (c *IndividualController) DeleteContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	contactIDStr := chi.URLParam(r, "contactId")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact ID")
		return
	}

	err = c.indService.DeleteContact(id, contactID)
	if errors.Is(err, services.ErrIndividualNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// GetSignals handles GET /individuals/{id}/signals
func (c *IndividualController) GetSignals(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	var signalType *string
	if st := r.URL.Query().Get("signal_type"); st != "" {
		signalType = &st
	}

	limit := lib.ParseIntQueryParam(r, "limit", 20)

	signals, err := c.indService.GetSignals(id, signalType, limit)
	if errors.Is(err, services.ErrIndividualNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, signals)
}

// CreateSignal handles POST /individuals/{id}/signals
func (c *IndividualController) CreateSignal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	var input schemas.SignalCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.indService.CreateSignal(id, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// DeleteSignal handles DELETE /individuals/{id}/signals/{signalId}
func (c *IndividualController) DeleteSignal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid individual ID")
		return
	}

	signalIDStr := chi.URLParam(r, "signalId")
	signalID, err := strconv.ParseInt(signalIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid signal ID")
		return
	}

	if err := c.indService.DeleteSignal(id, signalID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}
