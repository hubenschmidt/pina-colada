package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// ProposalController handles proposal HTTP requests
type ProposalController struct {
	proposalService *services.ProposalService
}

// NewProposalController creates a new proposal controller
func NewProposalController(proposalService *services.ProposalService) *ProposalController {
	return &ProposalController{proposalService: proposalService}
}

// GetPending handles GET /proposals
func (c *ProposalController) GetPending(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeProposalError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	result, err := c.proposalService.GetPendingProposals(tenantID, page, pageSize)
	if err != nil {
		writeProposalError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeProposalJSON(w, http.StatusOK, result)
}

// Approve handles POST /proposals/{id}/approve
func (c *ProposalController) Approve(w http.ResponseWriter, r *http.Request) {
	proposalID, err := parseProposalID(r)
	if err != nil {
		writeProposalError(w, http.StatusBadRequest, "invalid proposal ID")
		return
	}

	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeProposalError(w, http.StatusUnauthorized, "user not found")
		return
	}

	result, err := c.proposalService.ApproveProposal(proposalID, reviewerID)
	writeProposalResponse(w, http.StatusOK, result, err)
}

// Reject handles POST /proposals/{id}/reject
func (c *ProposalController) Reject(w http.ResponseWriter, r *http.Request) {
	proposalID, err := parseProposalID(r)
	if err != nil {
		writeProposalError(w, http.StatusBadRequest, "invalid proposal ID")
		return
	}

	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeProposalError(w, http.StatusUnauthorized, "user not found")
		return
	}

	result, err := c.proposalService.RejectProposal(proposalID, reviewerID)
	writeProposalResponse(w, http.StatusOK, result, err)
}

type bulkProposalRequest struct {
	IDs []int64 `json:"ids"`
}

// BulkApprove handles POST /proposals/bulk-approve
func (c *ProposalController) BulkApprove(w http.ResponseWriter, r *http.Request) {
	var req bulkProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProposalError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeProposalError(w, http.StatusUnauthorized, "user not found")
		return
	}

	succeeded, errs := c.proposalService.BulkApprove(req.IDs, reviewerID)
	writeBulkResponse(w, succeeded, errs, req.IDs)
}

// BulkReject handles POST /proposals/bulk-reject
func (c *ProposalController) BulkReject(w http.ResponseWriter, r *http.Request) {
	var req bulkProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProposalError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeProposalError(w, http.StatusUnauthorized, "user not found")
		return
	}

	succeeded, errs := c.proposalService.BulkReject(req.IDs, reviewerID)
	writeBulkResponse(w, succeeded, errs, req.IDs)
}

type updatePayloadRequest struct {
	Payload json.RawMessage `json:"payload"`
}

// UpdatePayload handles PUT /proposals/{id}/payload
func (c *ProposalController) UpdatePayload(w http.ResponseWriter, r *http.Request) {
	proposalID, err := parseProposalID(r)
	if err != nil {
		writeProposalError(w, http.StatusBadRequest, "invalid proposal ID")
		return
	}

	var req updatePayloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProposalError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.proposalService.UpdateProposalPayload(proposalID, req.Payload)
	writeProposalResponse(w, http.StatusOK, result, err)
}

func parseProposalID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func writeProposalResponse(w http.ResponseWriter, successStatus int, result *serializers.ProposalResponse, err error) {
	if errors.Is(err, services.ErrProposalNotFound) {
		writeProposalError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, services.ErrProposalNotPending) {
		writeProposalError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, services.ErrProposalHasValidationErrors) {
		writeProposalError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if errors.Is(err, services.ErrProposalExecutionFailed) {
		writeProposalError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err != nil {
		writeProposalError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProposalJSON(w, successStatus, result)
}

func writeBulkResponse(w http.ResponseWriter, succeeded []serializers.ProposalResponse, errs []error, requestedIDs []int64) {
	failed := make([]serializers.BulkError, 0, len(errs))
	successIDs := make(map[int64]bool)
	for _, s := range succeeded {
		successIDs[s.ID] = true
	}

	errIdx := 0
	for _, id := range requestedIDs {
		if successIDs[id] {
			continue
		}
		errMsg := "unknown error"
		if errIdx < len(errs) {
			errMsg = errs[errIdx].Error()
			errIdx++
		}
		failed = append(failed, serializers.BulkError{ID: id, Error: errMsg})
	}

	writeProposalJSON(w, http.StatusOK, serializers.BulkProposalResponse{
		Succeeded: succeeded,
		Failed:    failed,
	})
}

func writeProposalJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeProposalError(w http.ResponseWriter, status int, message string) {
	writeProposalJSON(w, status, serializers.ErrorResponse{Error: message})
}
