package allocation

import "time"

type Type string

const (
	TypeSNI   Type = "sni"
	TypeRoute Type = "route"
)

type Allocation struct {
	Domain string `json:"domain"`

	Type Type `json:"type"`

	// Address is the middle-server address that DNS should return
	// for a HIJACK allocation.
	Address string `json:"address,omitempty"`

	// SNI is the hostname/SNI allocation associated with this domain.
	SNI string `json:"sni,omitempty"`

	// RouteID identifies the routing allocation for ROUTE domains.
	RouteID string `json:"route_id,omitempty"`

	// AgentID identifies the agent responsible for this allocation.
	AgentID string `json:"agent_id"`

	// Revision changes whenever the allocation changes.
	Revision uint64 `json:"revision"`

	UpdatedAt time.Time `json:"updated_at"`
}
