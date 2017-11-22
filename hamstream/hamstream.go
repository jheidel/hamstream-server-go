package hamstream

import(
  "io"
  "fmt"
  "net/http"
	"hamstream/audio"
)


type Hamstream struct {
  Address string

  // quit signals to terminate
  quit chan bool
}

func NewServer(addr string) *Hamstream {
  return &Hamstream{
    Address: addr,
    quit: make(chan bool),
  }
}

func (h *Hamstream) loop() error {
    ai := audio.NewAudioInput()
    audioc, err := ai.Open()
    if err != nil {
      return err
    }
    defer ai.Close()

    bcast := audio.NewBroadcaster()
    aserver := audio.AudioServer{
      Broadcaster: bcast,
    }

    sfilter := audio.NewSilenceFilter()
    audioc = sfilter.Apply(audioc)
    bcast.Consume(audioc)

    http.HandleFunc("/audio", aserver.ServeAudio)
    http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
      io.WriteString(w, "OK\n")
    })

    errc := make(chan error)
    go func() {
      defer close(errc)
      fmt.Println("HTTP server listening on", h.Address)
      if err := http.ListenAndServe(h.Address, nil); err != nil {
        fmt.Println("Http serve error", err)
        errc <- err
      }
    }()

    select{
      case err := <-errc:
        return err
      case <-h.quit:
        // TODO quit the server.
        return nil
    }
}


func (h *Hamstream) Serve() <-chan error {
  errc := make(chan error)
  go func() {
    defer close(errc)
    if err := h.loop(); err != nil {
      errc <- err
    }
  }()
  return errc
}

func (h *Hamstream) Quit() {
  // Signal quit.
  h.quit <- true
}

