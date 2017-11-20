package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"hamstream/audio"
	"hamstream/flasher"
)

func main() {
	fmt.Println("hello world this is hamstream and I will flash led")

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Signal liveness by flashing an LED.
	flasher.Start()
	defer flasher.Stop()

	ai := audio.NewAudioInput()
	audioc, err := ai.Open()
	if err != nil {
		fmt.Printf("Audio exited with error %v\n", err)
	}
	defer ai.Close()

	for {
		select {
		case a := <-audioc:
			fmt.Printf("New audio chunk size %d: Sample %+v\n", len(a), a[:10])

		case s := <-sig:
			fmt.Printf("\nReceived signal %v. Exiting...\n", s)
			return
		}
	}
}
