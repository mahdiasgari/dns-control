package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	listen  string
	handler *Handler
}

func NewServer(
	listen string,
	handler *Handler,
) *Server {
	return &Server{
		listen:  listen,
		handler: handler,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	s.handler.Routes(mux)

	server := &http.Server{
		Addr:              s.listen,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api: listening on %s", s.listen)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("API server: %w", err)
	}

	return nil
}

func loggingMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			start := time.Now()

			next.ServeHTTP(w, r)

			log.Printf(
				"api: method=%s path=%s duration=%s",
				r.Method,
				r.URL.Path,
				time.Since(start),
			)
		},
	)
}
