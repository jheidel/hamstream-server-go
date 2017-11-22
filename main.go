package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"hamstream/flasher"
	"hamstream/hamstream"
)

func main() {
	fmt.Println("hello world this is hamstream and I will flash led")

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Signal liveness by flashing an LED.
	flasher.Start()
	defer flasher.Stop()

	h := hamstream.NewServer("localhost:8080")
	servec := h.Serve()

	for {
		select {
		case s := <-sig:
			fmt.Printf("\nReceived signal %v. Exiting...\n", s)
			h.Quit()
		case err := <-servec:
			if err != nil {
				fmt.Println("Hamstream exited with error ", err)
			}
			return
		}
	}
}
