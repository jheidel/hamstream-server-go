package main

import (
  "fmt"
  "os"
  "os/signal"
  "syscall"

  "hamstream/flasher"
  "hamstream/audio"
)

func main() {
	fmt.Println("hello world this is hamstream and I will flash led")

  sig := make(chan os.Signal)
  signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

  // Signal liveness by flashing an LED.
  flasher.Start()
  defer flasher.Stop()

  err := audio.Test()
  if err != nil {
    fmt.Printf("Audio exited with error %v\n", err)
  }

  // Wait for signal to stop program.
  s := <-sig
  fmt.Printf("\nReceived signal %v. Exiting...\n", s)
}
