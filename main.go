package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"hamstream-server-go/audio"
	"hamstream-server-go/flasher"
	"hamstream-server-go/server"
)

func main() {
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
		Path: "/tmp/out.wav",
	}

	if err := wr.OpenAndHost(ctx, &wg); err != nil {
		log.Fatalf("Failed to open output WAV: %v", err)
	}

	ai := audio.NewAudioInput()
	ai.Broadcaster = bcast
	ai.Filter = sf
	ai.Wav = wr
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
		Address:     ":8080",
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
