package allocation

import (
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

func normalize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	return strings.ToLower(value)
}

func (s *Store) EnsurePending(
	domain string,
	t Type,
) {
	domain = normalize(domain)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.leases[domain]; ok {
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

func (s *Store) Poll(
	agentID string,
	max int,
) []Lease {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	result := []Lease{}

	for _, lease := range s.leases {

		if lease.State == StateAssigned &&
			now.After(lease.ExpiresAt) {

			lease.State = StatePending
			lease.AgentID = ""
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

func (s *Store) Heartbeat(
	agentID string,
	leaseIDs []string,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for _, lease := range s.leases {

		if lease.AgentID != agentID {
			continue
		}

		for _, id := range leaseIDs {

			if lease.ID != id {
				continue
			}

			lease.ExpiresAt = now.Add(
				s.leaseDuration,
			)
		}
	}
}

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

		if lease.AgentID != agentID {
			return false
		}

		s.revision++

		lease.State = StateActive
		lease.Address = address
		lease.SNI = sni
		lease.RouteID = routeID
		lease.Revision = s.revision

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

	return lease, ok
}
