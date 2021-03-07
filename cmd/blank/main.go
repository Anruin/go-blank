package main

import (
	"context"
	"github.com/anruin/go-blank/pkg/monitoring"
	"github.com/anruin/go-blank/pkg/names"
	"github.com/anruin/go-blank/pkg/sync"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var err error

	// Initialize configuration.
	cfg := Config{}
	if ctx, err = cfg.Initialize(ctx); err != nil {
		log.Panicf("failed to initialize config: %v", err)
	}

	// Process flags.
	if err = processFlags(); err != nil {
		log.Panicf("failed to process flags: %v", err)
	}

	// Initialize sync.
	if ctx, err = sync.Initialize(ctx); err != nil {
		log.Panicf("failed to initialize sync: %v", err)
	}

	// Initialize monitoring.
	if ctx, err = monitoring.Initialize(ctx, cfg.Monitoring); err != nil {
		log.Panicf("failed to initialize monitoring: %v", err)
	}

	// Allow program goroutines to work until program is interrupted explicitly.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Infof("received an interrupt signal")

	// Cancel the main context to close all nested contexts.
	cancel()

	shutdown(cfg, ctx, err)
}

func shutdown(cfg Config, ctx context.Context, err error) {
	// Create a new context with timeout to wait for goroutines to shutdown gracefully.
	exitCtx, cancel := context.WithTimeout(context.Background(), time.Second*(time.Duration)(cfg.Shutdown.Timeout))

	// Run utility goroutine to exit by timeout.
	go func() {
		select {
		case <-exitCtx.Done():
			log.Infof("shutdown timeout")
			os.Exit(0)
		}
	}()

	// Wait for the error group to finish.
	errGroup := ctx.Value(names.CtxErrGroup).(*errgroup.Group)
	if err = errGroup.Wait(); err != nil {
		log.Errorf("failed to wait for error group goroutines to finish: %v", err)
	} else {
		log.Debugf("all goroutines exited")
	}

	// Cancel context if finished before timeout.
	cancel()
}
