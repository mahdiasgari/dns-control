package api

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	listen  string
	handler http.Handler
}

func NewServer(
	listen string,
	handler http.Handler,
) *Server {
	return &Server{
		listen:  listen,
		handler: handler,
	}
}

func (s *Server) Start() error {
	if s.listen == "" {
		return fmt.Errorf("API listen address is empty")
	}

	if s.handler == nil {
		return fmt.Errorf("API handler is nil")
	}

	server := &http.Server{
		Addr:    s.listen,
		Handler: s.handler,
	}

	log.Printf(
		"api: listening on %s",
		s.listen,
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		return fmt.Errorf("API server: %w", err)
	}

	return nil
}
