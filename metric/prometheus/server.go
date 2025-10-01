package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	currentServer *server
	serverMu      sync.Mutex
)

// server manages the HTTP server for Prometheus metrics
type server struct {
	httpServer *http.Server
	port       int
	path       string
	registry   *prometheus.Registry
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// newServer creates a new HTTP server for metrics
func newServer(port int, path string, registry *prometheus.Registry) *server {
	return &server{
		port:     port,
		path:     path,
		registry: registry,
		stopChan: make(chan struct{}),
	}
}

// start starts the HTTP server
func (s *server) start() error {
	mux := http.NewServeMux()
	
	// Register metrics handler
	mux.Handle(s.path, promhttp.HandlerFor(
		s.registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	
	// Register health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		logger().I("port", s.port).S("path", s.path).Info("starting HTTP server")
		
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger().I("port", s.port).Error(err)
		}
	}()
	
	return nil
}

// stop gracefully stops the HTTP server
func (s *server) stop() error {
	close(s.stopChan)
	
	if s.httpServer == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	
	s.wg.Wait()
	return nil
}

// startServer starts the global HTTP server
func startServer(port int, path string, registry *prometheus.Registry) error {
	serverMu.Lock()
	defer serverMu.Unlock()
	
	// Stop existing server if running
	if currentServer != nil {
		currentServer.stop()
	}
	
	currentServer = newServer(port, path, registry)
	return currentServer.start()
}

// stopServer stops the global HTTP server
func stopServer() error {
	serverMu.Lock()
	defer serverMu.Unlock()
	
	if currentServer == nil {
		return nil
	}
	
	err := currentServer.stop()
	currentServer = nil
	return err
}

// Close stops the HTTP server and cleans up resources
func Close() error {
	return stopServer()
}
