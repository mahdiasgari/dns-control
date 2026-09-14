package api

import "github.com/wraplink/dns-control/internal/allocation"

type AllocationRequest struct {
	Domain string `json:"domain"`

	Type allocation.Type `json:"type"`

	Address string `json:"address,omitempty"`

	SNI string `json:"sni,omitempty"`

	RouteID string `json:"route_id,omitempty"`

	AgentID string `json:"agent_id"`
}

type AllocationResponse struct {
	Allocation allocation.Allocation `json:"allocation"`
}

type PollResponse struct {
	Revision uint64 `json:"revision"`

	Allocations []allocation.Allocation `json:"allocations"`
}

type DeleteAllocationRequest struct {
	Domain string `json:"domain"`
}
