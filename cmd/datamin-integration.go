package main

import (
	"context"
	"integration/pkg/cli"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	time.Local = time.UTC

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		wait := make(chan os.Signal, 1)
		signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
		<-wait                                        // wait for SIGINT/SIGTERM
		signal.Reset(syscall.SIGINT, syscall.SIGTERM) // resetting signal listener, so that repeated Ctrl+C will exit immediately
		cancel()                                      // graceful stop
	}()

	err := cli.NewApplication().RunContext(ctx, os.Args)
	if err != nil {
		log.Fatalf("error running application: %v", err)
	} else {
		log.Info("Graceful shutdown")
	}
}
