package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/services"
)

type AccountController struct {
	accountService *services.AccountService
}

func NewAccountController(accountService *services.AccountService) *AccountController {
	return &AccountController{accountService: accountService}
}

func (c *AccountController) SearchAccounts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	results, err := c.accountService.SearchAccounts(query, tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// CreateRelationship handles POST /accounts/{id}/relationships
func (c *AccountController) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	accountID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	var input services.RelationshipCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.accountService.CreateRelationship(accountID, input, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteRelationship handles DELETE /accounts/{id}/relationships/{relationshipId}
func (c *AccountController) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	accountID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	relIDStr := chi.URLParam(r, "relationshipId")
	relationshipID, err := strconv.ParseInt(relIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid relationship ID")
		return
	}

	if err := c.accountService.DeleteRelationship(accountID, relationshipID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
