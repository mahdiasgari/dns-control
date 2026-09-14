package allocation

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex

	allocations map[string]Allocation

	revision uint64
}

func NewStore() *Store {
	return &Store{
		allocations: make(map[string]Allocation),
	}
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.TrimSuffix(domain, ".")
	return strings.ToLower(domain)
}

func (s *Store) Upsert(a Allocation) error {
	a.Domain = normalizeDomain(a.Domain)

	if a.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	if a.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}

	switch a.Type {
	case TypeSNI:
		if a.Address == "" {
			return fmt.Errorf("address is required for SNI allocation")
		}

	case TypeRoute:
		if a.Address == "" {
			return fmt.Errorf("address is required for ROUTE allocation")
		}

	default:
		return fmt.Errorf("invalid allocation type: %q", a.Type)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.revision++

	a.Revision = s.revision
	a.UpdatedAt = time.Now().UTC()

	s.allocations[a.Domain] = a

	return nil
}

func (s *Store) Delete(domain string) bool {
	domain = normalizeDomain(domain)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.allocations[domain]; !ok {
		return false
	}

	delete(s.allocations, domain)

	s.revision++

	return true
}

func (s *Store) Get(domain string) (Allocation, bool) {
	domain = normalizeDomain(domain)

	s.mu.RLock()
	defer s.mu.RUnlock()

	a, ok := s.allocations[domain]

	return a, ok
}

func (s *Store) List() []Allocation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Allocation, 0, len(s.allocations))

	for _, allocation := range s.allocations {
		result = append(result, allocation)
	}

	return result
}

func (s *Store) Revision() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.revision
}
