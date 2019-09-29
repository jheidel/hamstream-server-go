package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"hamstream-server-go/flasher"
	"hamstream-server-go/server"
)

func main() {
	log.Info("Start Hamstream Server")

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Signal liveness by flashing an LED.
	flasher.FlashAsync(ctx, &wg)

	errc := make(chan error)

	h := server.New(":8080")
	h.Serve(ctx, &wg, errc)

	for ctx.Err() == nil {
		select {
		case s := <-sig:
			log.Warnf("Received signal %q. Exiting...\n", s)
			cancel()
		case err := <-errc:
			if err != nil {
				log.Errorf("Runtime error: %v", err)
				cancel()
			}
		}
	}

	log.Info("Waiting for all contexts to exit")
	wg.Wait()
	log.Info("Hamstream Exit")
}
