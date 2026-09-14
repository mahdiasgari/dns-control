package dns

import (
	"log"
	"net"
	"strings"

	mdns "github.com/miekg/dns"

	"github.com/wraplink/dns-control/internal/allocation"
	"github.com/wraplink/dns-control/internal/domain"
)

type Handler struct {
	resolver    *Resolver
	domains     *domain.Store
	allocations *allocation.Store
	ttl         uint32
}

func NewHandler(
	resolver *Resolver,
	domains *domain.Store,
	allocations *allocation.Store,
	ttl uint32,
) *Handler {
	return &Handler{
		resolver:    resolver,
		domains:     domains,
		allocations: allocations,
		ttl:         ttl,
	}
}

func (h *Handler) ServeDNS(
	w mdns.ResponseWriter,
	req *mdns.Msg,
) {
	if req == nil || len(req.Question) == 0 {
		h.serverFailure(w, req)
		return
	}

	question := req.Question[0]

	domainName := normalizeDomain(question.Name)

	rule, matched := h.domains.Match(domainName)

	if !matched {
		h.handleNormal(w, req)
		return
	}

	switch rule.Mode {

	case domain.ModeNormal:
		h.handleNormal(w, req)

	case domain.ModeHijack:
		h.handleHijack(w, req, domainName)

	case domain.ModeRoute:
		h.handleRoute(w, req, domainName)

	default:
		log.Printf(
			"dns: unsupported mode=%q domain=%q",
			rule.Mode,
			domainName,
		)

		h.handleNormal(w, req)
	}
}

func (h *Handler) handleNormal(
	w mdns.ResponseWriter,
	req *mdns.Msg,
) {
	resp, err := h.resolver.Resolve(req)

	if err != nil {
		log.Printf("dns: upstream resolution failed: %v", err)

		h.serverFailure(w, req)
		return
	}

	if err := w.WriteMsg(resp); err != nil {
		log.Printf("dns: write normal response: %v", err)
	}
}

func (h *Handler) handleHijack(
	w mdns.ResponseWriter,
	req *mdns.Msg,
	domainName string,
) {
	/*
		HIJACK only changes A records.

		AAAA, CNAME, TXT, etc. are resolved normally.
	*/
	if req.Question[0].Qtype != mdns.TypeA {
		h.handleNormal(w, req)
		return
	}

	allocation, ok := h.allocations.Get(domainName)

	if !ok {
		log.Printf(
			"dns: no SNI allocation for %q",
			domainName,
		)

		h.handleNormal(w, req)
		return
	}

	ip := net.ParseIP(allocation.Address)

	if ip == nil || ip.To4() == nil {
		log.Printf(
			"dns: invalid SNI allocation address domain=%q address=%q",
			domainName,
			allocation.Address,
		)

		h.handleNormal(w, req)
		return
	}

	resp := new(mdns.Msg)
	resp.SetReply(req)

	resp.Answer = []mdns.RR{
		&mdns.A{
			Hdr: mdns.RR_Header{
				Name:   req.Question[0].Name,
				Rrtype: mdns.TypeA,
				Class:  mdns.ClassINET,
				Ttl:    h.ttl,
			},
			A: ip.To4(),
		},
	}

	if err := w.WriteMsg(resp); err != nil {
		log.Printf("dns: write hijack response: %v", err)
	}
}

func (h *Handler) handleRoute(
	w mdns.ResponseWriter,
	req *mdns.Msg,
	domainName string,
) {
	/*
		ROUTE is different from HIJACK.

		The DNS server does not itself transport UDP traffic.

		The agent allocates a routing endpoint, while the
		network/route component handles the actual traffic.

		For DNS, we currently return the allocated routing
		endpoint for A queries.
	*/
	if req.Question[0].Qtype != mdns.TypeA {
		h.handleNormal(w, req)
		return
	}

	allocation, ok := h.allocations.Get(domainName)

	if !ok {
		log.Printf(
			"dns: no ROUTE allocation for %q",
			domainName,
		)

		h.handleNormal(w, req)
		return
	}

	ip := net.ParseIP(allocation.Address)

	if ip == nil || ip.To4() == nil {
		log.Printf(
			"dns: invalid ROUTE address domain=%q address=%q",
			domainName,
			allocation.Address,
		)

		h.handleNormal(w, req)
		return
	}

	resp := new(mdns.Msg)
	resp.SetReply(req)

	resp.Answer = []mdns.RR{
		&mdns.A{
			Hdr: mdns.RR_Header{
				Name:   req.Question[0].Name,
				Rrtype: mdns.TypeA,
				Class:  mdns.ClassINET,
				Ttl:    h.ttl,
			},
			A: ip.To4(),
		},
	}

	if err := w.WriteMsg(resp); err != nil {
		log.Printf("dns: write route response: %v", err)
	}
}

func (h *Handler) serverFailure(
	w mdns.ResponseWriter,
	req *mdns.Msg,
) {
	resp := new(mdns.Msg)

	if req != nil {
		resp.SetRcode(req, mdns.RcodeServerFailure)
	} else {
		resp.SetRcode(&mdns.Msg{}, mdns.RcodeServerFailure)
	}

	if err := w.WriteMsg(resp); err != nil {
		log.Printf("dns: write SERVFAIL: %v", err)
	}
}

func normalizeDomain(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	return strings.ToLower(value)
}
