package flasher

import (
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"time"
)

var (
	// pin is a GPIO output connected to an LED
	pin = rpio.Pin(25)

	// done signals intent and acknowledgement of stopping
	done = make(chan chan bool)
)

func loop() {
	tick := time.Tick(time.Second / 4)
	for {
		select {
		case <-tick:
			pin.Toggle()
		case d := <-done:
			d <- true
			return
		}
	}
}

// Start begins flashing asynchronously.
func Start() {
	if err := rpio.Open(); err != nil {
		panic(fmt.Sprintf("unable to open rpio: %v", err))
	}
	pin.Output()
	go loop()
}

// Stop terminates the flashing.
func Stop() {
	d := make(chan bool)
	// Signal intent to stop.
	done <- d
	// Wait for acknowledgement of stop.
	<-d
	// Clean up.
	pin.Low()
	rpio.Close()
}
