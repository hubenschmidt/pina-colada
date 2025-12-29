package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"agent/internal/lib"
	"agent/internal/middleware"
	"agent/internal/schemas"
	"agent/internal/serializers"
	"agent/internal/services"
)

// OrganizationController handles organization HTTP requests
type OrganizationController struct {
	orgService *services.OrganizationService
}

// NewOrganizationController creates a new organization controller
func NewOrganizationController(orgService *services.OrganizationService) *OrganizationController {
	return &OrganizationController{orgService: orgService}
}

// GetOrganizations handles GET /organizations
func (c *OrganizationController) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	orderBy := r.URL.Query().Get("orderBy")
	order := r.URL.Query().Get("order")
	search := r.URL.Query().Get("search")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	result, err := c.orgService.GetOrganizations(page, limit, orderBy, order, search, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetOrganization handles GET /organizations/{id}
func (c *OrganizationController) GetOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	result, err := c.orgService.GetOrganization(id)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SearchOrganizations handles GET /organizations/search
func (c *OrganizationController) SearchOrganizations(w http.ResponseWriter, r *http.Request) {
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

	results, err := c.orgService.SearchOrganizations(query, tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// DeleteOrganization handles DELETE /organizations/{id}
func (c *OrganizationController) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	err = c.orgService.DeleteOrganization(id)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// UpdateOrganization handles PUT /organizations/{id}
func (c *OrganizationController) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var input schemas.OrganizationUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.orgService.UpdateOrganization(id, input, userID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// AddOrganizationContact handles POST /organizations/{id}/contacts
func (c *OrganizationController) AddOrganizationContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var input schemas.OrgContactCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())

	result, err := c.orgService.AddContactToOrganization(orgID, input, userID, tenantID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// CreateOrganization handles POST /organizations
func (c *OrganizationController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var input schemas.OrganizationCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}
	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.orgService.CreateOrganization(input, userID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetOrganizationContacts handles GET /organizations/{id}/contacts
func (c *OrganizationController) GetOrganizationContacts(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	result, err := c.orgService.GetOrganizationContacts(orgID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateOrganizationContact handles PUT /organizations/{id}/contacts/{contactId}
func (c *OrganizationController) UpdateOrganizationContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	contactIDStr := chi.URLParam(r, "contactId")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact ID")
		return
	}

	var input schemas.OrgContactUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.orgService.UpdateOrganizationContact(orgID, contactID, input, userID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteOrganizationContact handles DELETE /organizations/{id}/contacts/{contactId}
func (c *OrganizationController) DeleteOrganizationContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	contactIDStr := chi.URLParam(r, "contactId")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact ID")
		return
	}

	err = c.orgService.DeleteOrganizationContact(orgID, contactID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// GetOrganizationTechnologies handles GET /organizations/{id}/technologies
func (c *OrganizationController) GetOrganizationTechnologies(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	result, err := c.orgService.GetOrganizationTechnologies(orgID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// AddOrganizationTechnology handles POST /organizations/{id}/technologies
func (c *OrganizationController) AddOrganizationTechnology(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var input schemas.OrgTechnologyCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = c.orgService.AddTechnologyToOrganization(orgID, input)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, serializers.SuccessResponse{Success: true})
}

// RemoveOrganizationTechnology handles DELETE /organizations/{id}/technologies/{techId}
func (c *OrganizationController) RemoveOrganizationTechnology(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	techIDStr := chi.URLParam(r, "techId")
	techID, err := strconv.ParseInt(techIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid technology ID")
		return
	}

	if err := c.orgService.RemoveTechnologyFromOrganization(orgID, techID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// GetOrganizationFundingRounds handles GET /organizations/{id}/funding-rounds
func (c *OrganizationController) GetOrganizationFundingRounds(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	result, err := c.orgService.GetOrganizationFundingRounds(orgID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateOrganizationFundingRound handles POST /organizations/{id}/funding-rounds
func (c *OrganizationController) CreateOrganizationFundingRound(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var input schemas.FundingRoundCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.orgService.CreateOrganizationFundingRound(orgID, input)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// DeleteOrganizationFundingRound handles DELETE /organizations/{id}/funding-rounds/{roundId}
func (c *OrganizationController) DeleteOrganizationFundingRound(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	roundIDStr := chi.URLParam(r, "roundId")
	roundID, err := strconv.ParseInt(roundIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid round ID")
		return
	}

	if err := c.orgService.DeleteOrganizationFundingRound(orgID, roundID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// GetOrganizationSignals handles GET /organizations/{id}/signals
func (c *OrganizationController) GetOrganizationSignals(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var signalType *string
	if st := r.URL.Query().Get("signal_type"); st != "" {
		signalType = &st
	}

	limit := lib.ParseIntQueryParam(r, "limit", 20)

	result, err := c.orgService.GetOrganizationSignals(orgID, signalType, limit)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateOrganizationSignal handles POST /organizations/{id}/signals
func (c *OrganizationController) CreateOrganizationSignal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var input schemas.SignalCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.orgService.CreateOrganizationSignal(orgID, input, userID)
	if errors.Is(err, services.ErrOrganizationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// DeleteOrganizationSignal handles DELETE /organizations/{id}/signals/{signalId}
func (c *OrganizationController) DeleteOrganizationSignal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	signalIDStr := chi.URLParam(r, "signalId")
	signalID, err := strconv.ParseInt(signalIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid signal ID")
		return
	}

	if err := c.orgService.DeleteOrganizationSignal(orgID, signalID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}
