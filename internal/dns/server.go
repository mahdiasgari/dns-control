package dns

import (
	"fmt"
	"log"

	mdns "github.com/miekg/dns"
)

type Server struct {
	listen  string
	handler mdns.Handler
}

func NewServer(
	listen string,
	handler mdns.Handler,
) *Server {
	return &Server{
		listen:  listen,
		handler: handler,
	}
}

func (s *Server) Start() error {
	if s.listen == "" {
		return fmt.Errorf("DNS listen address is empty")
	}

	udp := &mdns.Server{
		Addr:    s.listen,
		Net:     "udp",
		Handler: s.handler,
	}

	tcp := &mdns.Server{
		Addr:    s.listen,
		Net:     "tcp",
		Handler: s.handler,
	}

	errCh := make(chan error, 2)

	go func() {
		log.Printf("dns: UDP listening on %s", s.listen)

		if err := udp.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("udp DNS: %w", err)
		}
	}()

	go func() {
		log.Printf("dns: TCP listening on %s", s.listen)

		if err := tcp.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("tcp DNS: %w", err)
		}
	}()

	return <-errCh
}
