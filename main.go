package main

import (
	"context"
	"kafkatrigger/cli"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	var err error
	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	ctx, cancel := context.WithCancel(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		defer close(done)
		err = cli.NewApplication().RunContext(ctx, os.Args)
		if err != nil {
			log.Fatalf("error running application: %v", err)
		} else {
			log.Info("Graceful shutdown")
		}
	}()

	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-wait: // wait for SIGINT/SIGTERM
		signal.Reset(syscall.SIGINT, syscall.SIGTERM) // resetting signal listener, so that repeated Ctrl+C will exit immediately
		cancel()                                      // graceful stop
		<-done

	case <-done:
		cancel() // graceful stop
	}
}
