package monitoring

import (
	"context"
	"fmt"
	"github.com/anruin/go-blank/pkg/names"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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

	if g, ok := ctx.Value(names.CtxErrGroup).(*errgroup.Group); !ok {
		log.Panicf("failed to get errgroup")
	} else {
		g.Go(func() error {
			log.Debugf("monitoring #1")
			log.Infof("monitoring server listens at: %s", s.Http.Addr)
			if err := s.Http.ListenAndServe(); err != http.ErrServerClosed {
				log.Errorf("failed to start monitoring server: %v", err)
				return err
			} else {
				log.Infof("monitoring server closed: %v", err)
			}
			return nil
		})

		// #2: Utility goroutine to wait for the main context cancel.
		g.Go(func() error {
			log.Debugf("monitoring #2")
			select {
			case <-ctx.Done(): // #1 Exit
				log.Debugf("closing monitoring server")

				t := time.Duration(cfg.Timeout) * time.Second

				// Wait for graceful shutdown.
				c, cancel := context.WithTimeout(context.Background(), t)
				if err := s.Http.Shutdown(c); err != nil {
					log.Errorf("failed to shutdown server: %v", err)
				} else {
					cancel()
				}

				// Wait for the server shutdown and quit the utility goroutine.
				select {
				case <-c.Done():
					log.Infof("closed monitoring server")
				case <-time.After(t):
					log.Errorf("monitoring server close timeout")
				}

				cancel() // #2 Exit
			}

			return nil
		})
	}

	return ctx, nil
}
