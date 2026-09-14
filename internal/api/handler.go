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
	token       string
}

func NewHandler(
	domains *domain.Store,
	allocations *allocation.Store,
	token string,
) http.Handler {
	h := &Handler{
		domains:     domains,
		allocations: allocations,
		token:       token,
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/v1/agent/register",
		h.register,
	)

	mux.HandleFunc(
		"/api/v1/agent/poll",
		h.poll,
	)

	mux.HandleFunc(
		"/api/v1/agent/heartbeat",
		h.heartbeat,
	)

	mux.HandleFunc(
		"/api/v1/agent/report",
		h.report,
	)

	mux.HandleFunc(
		"/api/v1/domains",
		h.domainsList,
	)

	return h.auth(mux)
}

func (h *Handler) auth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if h.token != "" {
			auth := r.Header.Get("Authorization")

			expected := "Bearer " + h.token

			if auth != expected {
				writeJSON(
					w,
					http.StatusUnauthorized,
					map[string]any{
						"error": "unauthorized",
					},
				)

				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) register(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "invalid request body",
			},
		)
		return
	}

	req.AgentID = strings.TrimSpace(req.AgentID)

	if req.AgentID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "agent_id is required",
			},
		)
		return
	}

	// Registration/capacity tracking will be added to the
	// allocation scheduler. For now, acknowledge the agent.
	writeJSON(
		w,
		http.StatusOK,
		RegisterResponse{
			OK: true,
		},
	)
}

func (h *Handler) poll(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req PollRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "invalid request body",
			},
		)
		return
	}

	req.AgentID = strings.TrimSpace(req.AgentID)

	if req.AgentID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "agent_id is required",
			},
		)
		return
	}

	if req.Max <= 0 {
		req.Max = 1
	}

	leases := h.allocations.Poll(
		req.AgentID,
		req.Max,
	)

	writeJSON(
		w,
		http.StatusOK,
		PollResponse{
			Leases: leases,
		},
	)
}

func (h *Handler) heartbeat(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req HeartbeatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "invalid request body",
			},
		)
		return
	}

	req.AgentID = strings.TrimSpace(req.AgentID)

	if req.AgentID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "agent_id is required",
			},
		)
		return
	}

	h.allocations.Heartbeat(
		req.AgentID,
		req.LeaseIDs,
	)

	writeJSON(
		w,
		http.StatusOK,
		HeartbeatResponse{
			OK: true,
		},
	)
}

func (h *Handler) report(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req ReportRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "invalid request body",
			},
		)
		return
	}

	req.AgentID = strings.TrimSpace(req.AgentID)
	req.LeaseID = strings.TrimSpace(req.LeaseID)

	if req.AgentID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "agent_id is required",
			},
		)
		return
	}

	if req.LeaseID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"error": "lease_id is required",
			},
		)
		return
	}

	if !req.Success {
		// Failure handling will be added to Store.
		writeJSON(
			w,
			http.StatusOK,
			ReportResponse{
				OK: true,
			},
		)

		return
	}

	ok := h.allocations.Activate(
		req.LeaseID,
		req.AgentID,
		req.Address,
		req.SNI,
		req.RouteID,
	)

	if !ok {
		writeJSON(
			w,
			http.StatusConflict,
			map[string]any{
				"error": "lease activation rejected",
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		ReportResponse{
			OK: true,
		},
	)
}

func (h *Handler) domainsList(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		h.domains.List(),
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}
