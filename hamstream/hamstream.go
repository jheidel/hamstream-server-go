package hamstream

import(
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

    errc := make(chan error)
    go func() {
      defer close(errc)
      if err := http.ListenAndServe(h.Address, nil); err != nil {
        errc <- err
      }
    }()

    select{
      case err := <-errc:
        return err
      case <-h.quit:
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

