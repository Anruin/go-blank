package monitoring

import (
	"context"
	"github.com/anruin/go-blank/pkg/names"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	StatusOk    = "ok"
	StatusError = "error"
)

type Server struct {
	Router *mux.Router
	Http   *http.Server
	status string
}

// Initialize routes.
func (s *Server) initializeRoutes() {
	s.Router.Path("/").
		Methods("GET").
		HandlerFunc(s.handleIndex)
}

// HTTP server func.
func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(fmt.Sprintf(`{"status": "%s"}`, s.status))); err != nil {
		log.Errorf("failed to respond to monitoring request: %v", err)
	}
}

// Use to update the application status for monitoring.
func (s *Server) SetStatus(status string) {
	if s.status != status {
		log.Infof("status changed: %s", status)
		s.status = status
	}
}

func Initialize(ctx context.Context, cfg Config) (context.Context, error) {
	router := mux.NewRouter()

	// Create a monitoring server.
	s := &Server{
		status: StatusOk,
		Router: router,
		Http: &http.Server{
			Handler:      router,
			Addr:         net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
			WriteTimeout: 1 * time.Second,
			ReadTimeout:  1 * time.Second,
		},
	}

	// Add the monitoring server to the context.
	ctx = context.WithValue(ctx, names.CtxMonitoring, s)

	// #1: Main monitoring server goroutine.
	go func() {
		log.Infof("monitoring #1")
		log.Infof("monitoring server listens at: %s", s.Http.Addr)
		if err := s.Http.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("failed to start monitoring server: %v", err)
		} else {
			log.Infof("monitoring server closed: %v", err)
		}
	}()

	// #2: Utility goroutine to wait for the main context cancel.
	go func() {
		log.Infof("monitoring #2")
		select {
		case <-ctx.Done(): // #1 Exit
			log.Infof("closing monitoring server")

			t := 3 * time.Second

			// Wait for graceful shutdown.
			c, cancel := context.WithTimeout(context.Background(), t)
			if err := s.Http.Shutdown(c); err != nil {
				log.Errorf("failed to shutdown server: %v", err)
			}

			// Wait for the server shutdown and quit the utility goroutine.
			select {
			case <-c.Done():
				cancel() // #2 Exit
			case <-time.After(t):
				cancel() // #2 Exit
			}
		}
	}()

	return ctx, nil
}
