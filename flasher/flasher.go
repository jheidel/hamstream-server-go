package flasher

import (
	"fmt"
  log "github.com/sirupsen/logrus"
  "context"
	"github.com/stianeikeland/go-rpio"
  "sync"
	"time"
)

var (
	// pin is a GPIO output connected to an LED
	pin = rpio.Pin(25)
)

// Start begins flashing asynchronously until cancalled.
func FlashAsync(ctx context.Context, wg *sync.WaitGroup) {
  wg.Add(1)

	if err := rpio.Open(); err != nil {
		panic(fmt.Sprintf("unable to open rpio: %v", err))
	}
	pin.Output()

  go func() {
    defer wg.Done()
    log.Info("Starting periodic status flash")

    tick := time.Tick(time.Second / 4)
    for ctx.Err() == nil {
      select {
      case <-tick:
        pin.Toggle()
      case <-ctx.Done():
      }
    }

    log.Info("Stopping periodic status flash")
    pin.Low()
    rpio.Close()
  }()
}
