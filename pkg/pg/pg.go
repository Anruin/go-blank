package pg

import (
	"context"
	"fmt"
	"github.com/anruin/go-blank/pkg/monitoring"
	"github.com/anruin/go-blank/pkg/names"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func Initialize(ctx context.Context, cfg Config) (context.Context, error) {
	// Update monitoring status.
	if handler, ok := ctx.Value(names.CtxMonitoring).(*monitoring.Server); ok {
		handler.SetStatus(monitoring.StatusOk)
	}

	log.Debugf("connecting to the database")

	// Open connection to the database;
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name)
	if conn, err := sqlx.Open("postgres", connStr); err != nil {
		log.Errorf("failed to connect to the database: %v", err)
		// Update monitoring status.
		if handler, ok := ctx.Value(names.CtxMonitoring).(*monitoring.Server); ok {
			handler.SetStatus(monitoring.StatusError)
		}
		return nil, err
	} else {

		// Ping database to ensure successful connect.
		if err = conn.Ping(); err != nil {
			log.Errorf("failed to ping the database: %v", err)
			// Update monitoring status.
			if handler, ok := ctx.Value(names.CtxMonitoring).(*monitoring.Server); ok {
				handler.SetStatus(monitoring.StatusError)
			}
			return nil, err
		}

		// Utility goroutine to close the database connection at context cancel.
		if g, ok := ctx.Value(names.CtxErrGroup).(*errgroup.Group); !ok {
			log.Panicf("failed to get errgroup")
		} else {
			log.Tracef("go pg #1 enter")
			// #1: Database context goroutine.
			g.Go(func() error {
				select {
				case <-ctx.Done():
					if err := CloseConnection(ctx); err != nil {
						log.Errorf("failed to close database connection: %v", err)
						// Update monitoring status.
						if handler, ok := ctx.Value(names.CtxMonitoring).(*monitoring.Server); ok {
							handler.SetStatus(monitoring.StatusError)
						}
						log.Tracef("go monitoring #1 exit")
						return err
					}
				}
				log.Tracef("go monitoring #1 exit")
				return nil
			})
		}

		return context.WithValue(ctx, names.CtxPgConn, conn), nil
	}
}

func CloseConnection(ctx context.Context) error {
	log.Debugf("closing database connection")
	if db, ok := ctx.Value(names.CtxPgConn).(*sqlx.DB); ok {
		return db.Close()
	}
	return nil
}
