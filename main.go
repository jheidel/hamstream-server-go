package main

import (
	"context"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"hamstream-server-go/audio"
	"hamstream-server-go/flasher"
	"hamstream-server-go/server"
)

var (
	path = flag.String("path", "/home/pi/recordings/", "Path for output files")
	port = flag.Int("port", 8080, "Port to host web server")
)

func main() {
	flag.Parse()
	log.Info("Start Hamstream Server")

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	errc := make(chan error)

	// Signal liveness by flashing an LED.
	flasher.FlashAsync(ctx, &wg)

	bcast := audio.NewBroadcaster()
	sf := audio.NewSilenceFilter()

	wr := &audio.WavWriter{
		PathBase: *path,
	}

	if err := wr.OpenAndHost(ctx, &wg); err != nil {
		log.Fatalf("Failed to open output WAV: %v", err)
	}

	ai := &audio.AudioInput{
		Broadcaster: bcast,
		Filter:      sf,
		Wav:         wr,
	}
	if err := ai.OpenAndServe(ctx, &wg); err != nil {
		log.Fatalf("Failed to open audio source: %v", err)
	}

	aserver := &audio.AudioServer{
		Broadcaster: bcast,
	}

	sserver := &audio.StatsServer{
		Broadcaster: bcast,
		Filter:      sf,
		Input:       ai,
	}

	h := &server.Server{
		Address:     fmt.Sprintf(":%d", *port),
		AudioServer: aserver,
		StatsServer: sserver,
	}
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
