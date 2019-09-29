package server

import (
	"fmt"
	"hamstream-server-go/audio"
	"io"
	"net/http"
)

type Server struct {
	Address string

	// quit signals to terminate
	quit chan bool
}

func New(addr string) *Server {
	return &Server{
		Address: addr,
		quit:    make(chan bool),
	}
}

func (h *Server) loop() error {
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

	select {
	case err := <-errc:
		return err
	case <-h.quit:
		// TODO quit the server.
		return nil
	}
}

func (h *Server) Serve() <-chan error {
	errc := make(chan error)
	go func() {
		defer close(errc)
		if err := h.loop(); err != nil {
			errc <- err
		}
	}()
	return errc
}

func (h *Server) Quit() {
	// Signal quit.
	h.quit <- true
}
