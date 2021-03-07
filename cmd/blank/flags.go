package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
)

func processFlags() (err error) {

	verbose := flag.Bool("v", false, "verbose mode")
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	return nil
}
