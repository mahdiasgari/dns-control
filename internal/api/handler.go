package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/wraplink/dns-control/internal/allocation"
	"github.com/wraplink/dns-control/internal/domain"
)

type Handler struct {
	domains     *domain.Store
	allocations *allocation.Store
	agentToken  string
}

func NewHandler(
	domains *domain.Store,
	allocations *allocation.Store,
	agentToken string,
) *Handler {
	return &Handler{
		domains:     domains,
		allocations: allocations,
		agentToken:  agentToken,
	}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.health)

	mux.HandleFunc("/api/v1/domains", h.listDomains)

	mux.HandleFunc("/api/v1/allocations", h.allocationsEndpoint)

	mux.HandleFunc("/api/v1/agent/poll", h.agentPoll)
}

func (h *Handler) health(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (h *Handler) listDomains(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(w, http.StatusOK, map[string]any{
		"domains": h.domains.List(),
	})
}

func (h *Handler) allocationsEndpoint(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !h.authorize(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "unauthorized",
		})
		return
	}

	switch r.Method {

	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"allocations": h.allocations.List(),
		})

	case http.MethodPost:
		h.upsertAllocation(w, r)

	case http.MethodDelete:
		h.deleteAllocation(w, r)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) upsertAllocation(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request AllocationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid JSON",
		})
		return
	}

	a := allocation.Allocation{
		Domain:  request.Domain,
		Type:    request.Type,
		Address: request.Address,
		SNI:     request.SNI,
		RouteID: request.RouteID,
		AgentID: request.AgentID,
	}

	if err := h.allocations.Upsert(a); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	result, _ := h.allocations.Get(request.Domain)

	writeJSON(w, http.StatusOK, AllocationResponse{
		Allocation: result,
	})
}

func (h *Handler) deleteAllocation(
	w http.ResponseWriter,
	r *http.Request,
) {
	domain := strings.TrimSpace(
		r.URL.Query().Get("domain"),
	)

	if domain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "domain is required",
		})
		return
	}

	deleted := h.allocations.Delete(domain)

	writeJSON(w, http.StatusOK, map[string]any{
		"deleted": deleted,
	})
}

func (h *Handler) agentPoll(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !h.authorize(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "unauthorized",
		})
		return
	}

	writeJSON(w, http.StatusOK, PollResponse{
		Revision:    h.allocations.Revision(),
		Allocations: h.allocations.List(),
	})
}

func (h *Handler) authorize(
	r *http.Request,
) bool {
	if h.agentToken == "" {
		return true
	}

	header := r.Header.Get("Authorization")

	expected := "Bearer " + h.agentToken

	return header == expected
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
