package main

import (
	"context"
	"github.com/anruin/go-blank/pkg/monitoring"
	log "github.com/sirupsen/logrus"
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

	// Create a new context with timeout to wait for goroutines to shutdown gracefully.
	q, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	select {
	case <-q.Done():
		log.Infof("quit")
	}
}
