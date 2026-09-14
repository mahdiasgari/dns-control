package allocation

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	mu sync.RWMutex

	leases map[string]*Lease

	revision uint64

	leaseDuration time.Duration
}

func NewStore() *Store {
	return &Store{
		leases:        make(map[string]*Lease),
		leaseDuration: 30 * time.Second,
	}
}

// List returns a snapshot of all leases.
//
// The returned slice contains copies, so callers cannot modify
// the store without going through Store methods.
func (s *Store) List() []Lease {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Lease, 0, len(s.leases))

	for _, lease := range s.leases {
		result = append(result, *lease)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Domain < result[j].Domain
	})

	return result
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	return strings.ToLower(value)
}

// EnsurePending creates a pending lease for a domain if one does not
// already exist.
//
// The domain is the unique key for now.
func (s *Store) EnsurePending(
	domain string,
	t Type,
) {
	domain = normalize(domain)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.leases[domain]; exists {
		return
	}

	s.revision++

	s.leases[domain] = &Lease{
		ID:       uuid.NewString(),
		Domain:   domain,
		Type:     t,
		State:    StatePending,
		Revision: s.revision,
	}
}

// Poll assigns pending leases to an agent.
//
// ASSIGNED leases whose lease duration has expired are returned
// to PENDING first, allowing another agent to acquire them.
func (s *Store) Poll(
	agentID string,
	max int,
) []Lease {
	if max <= 0 {
		return nil
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]Lease, 0, max)

	for _, lease := range s.leases {

		// Reclaim expired assignments.
		if lease.State == StateAssigned &&
			!lease.ExpiresAt.IsZero() &&
			now.After(lease.ExpiresAt) {

			lease.State = StatePending
			lease.AgentID = ""
			lease.ExpiresAt = time.Time{}

			s.revision++
			lease.Revision = s.revision
		}

		if lease.State != StatePending {
			continue
		}

		s.revision++

		lease.State = StateAssigned
		lease.AgentID = agentID
		lease.Revision = s.revision
		lease.ExpiresAt = now.Add(s.leaseDuration)

		result = append(result, *lease)

		if len(result) >= max {
			break
		}
	}

	return result
}

// Heartbeat renews the lease expiration for leases owned by agentID.
//
// Unknown lease IDs are ignored.
// Leases belonging to another agent are ignored.
func (s *Store) Heartbeat(
	agentID string,
	leaseIDs []string,
) {
	if agentID == "" || len(leaseIDs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(s.leaseDuration)

	for _, lease := range s.leases {

		if lease.AgentID != agentID {
			continue
		}

		for _, id := range leaseIDs {

			if lease.ID != id {
				continue
			}

			lease.ExpiresAt = expiresAt

			break
		}
	}
}

// Activate changes an assigned lease to ACTIVE after the agent
// successfully configured its dataplane.
func (s *Store) Activate(
	leaseID string,
	agentID string,
	address string,
	sni string,
	routeID string,
) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, lease := range s.leases {

		if lease.ID != leaseID {
			continue
		}

		// Only the agent that owns the assignment may activate it.
		if lease.AgentID != agentID {
			return false
		}

		// Only ASSIGNED leases can become ACTIVE.
		if lease.State != StateAssigned {
			return false
		}

		s.revision++

		lease.State = StateActive
		lease.Address = address
		lease.SNI = sni
		lease.RouteID = routeID
		lease.Revision = s.revision

		// ACTIVE leases still need to be renewed by heartbeat.
		lease.ExpiresAt = time.Now().Add(s.leaseDuration)

		return true
	}

	return false
}

func (s *Store) Get(
	domain string,
) (*Lease, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	domain = normalize(domain)

	lease, ok := s.leases[domain]

	if !ok {
		return nil, false
	}

	// Return a copy so the caller cannot modify internal state.
	copy := *lease

	return &copy, true
}

func (s *Store) Revision() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.revision
}
