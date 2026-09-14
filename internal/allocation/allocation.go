package allocation

import "time"

type State string

const (
	StatePending  State = "pending"
	StateAssigned State = "assigned"
	StateActive   State = "active"
	StateFailed   State = "failed"
)

type Type string

const (
	TypeSNI   Type = "sni"
	TypeRoute Type = "route"
)

type Lease struct {
	ID string `json:"id"`

	Domain string `json:"domain"`

	Type Type `json:"type"`

	State State `json:"state"`

	AgentID string `json:"agent_id,omitempty"`

	Address string `json:"address,omitempty"`

	SNI string `json:"sni,omitempty"`

	RouteID string `json:"route_id,omitempty"`

	Revision uint64 `json:"revision"`

	ExpiresAt time.Time `json:"expires_at"`
}
