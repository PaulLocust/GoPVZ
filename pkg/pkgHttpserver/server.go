// Package httpserver implements HTTP server.
package pkgHttpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	_defaultAddr            = ":80"
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultShutdownTimeout = 3 * time.Second
)

// Server -.
type Server struct {
	server          *http.Server
	router         *gin.Engine
	notify         chan error
	address        string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	shutdownTimeout time.Duration
}

// New -.
func New(opts ...Option) *Server {
	router := gin.New()

	s := &Server{
		router:         router,
		notify:         make(chan error, 1),
		address:        _defaultAddr,
		readTimeout:    _defaultReadTimeout,
		writeTimeout:   _defaultWriteTimeout,
		shutdownTimeout: _defaultShutdownTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	s.server = &http.Server{
		Addr:         s.address,
		Handler:      router,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
	}

	return s
}

// Start -.
func (s *Server) Start() {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.notify <- err
		}
		close(s.notify)
	}()
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// GetRouter возвращает экземпляр gin.Engine для настройки маршрутов
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}