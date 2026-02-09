// Package server provides HTTP server implementation.
package httpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Krokozabra213/effective_mobile/internal/config"
)

// Server wraps HTTP server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new HTTP server with the given config and handler.
func NewServer(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
			Handler:        handler,
			ReadTimeout:    cfg.HTTP.ReadTimeout,
			WriteTimeout:   cfg.HTTP.WriteTimeout,
			MaxHeaderBytes: cfg.HTTP.MaxHeaderMegabytes << 20,
		},
	}
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

// ShutDown gracefully stops the server with the given timeout.
func (s *Server) ShutDown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
