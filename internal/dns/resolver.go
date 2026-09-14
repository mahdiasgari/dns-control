package dns

import (
	"fmt"
	"log"
	"time"

	mdns "github.com/miekg/dns"
)

type Resolver struct {
	client    *mdns.Client
	upstreams []string
}

func NewResolver(upstreams []string) *Resolver {
	return &Resolver{
		client: &mdns.Client{
			Timeout: 5 * time.Second,
		},
		upstreams: append([]string(nil), upstreams...),
	}
}

func (r *Resolver) Resolve(req *mdns.Msg) (*mdns.Msg, error) {
	if req == nil {
		return nil, fmt.Errorf("nil DNS request")
	}

	if len(req.Question) == 0 {
		return nil, fmt.Errorf("DNS request has no questions")
	}

	if len(r.upstreams) == 0 {
		return nil, fmt.Errorf("no upstream DNS servers configured")
	}

	var lastErr error

	for _, upstream := range r.upstreams {
		resp, _, err := r.client.Exchange(req, upstream)

		if err == nil {
			return resp, nil
		}

		lastErr = err

		log.Printf(
			"dns: upstream %s failed: %v",
			upstream,
			err,
		)
	}

	return nil, fmt.Errorf(
		"all upstream DNS servers failed: %w",
		lastErr,
	)
}
