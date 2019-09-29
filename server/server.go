package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"hamstream-server-go/audio"
	"io"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	Address string
}

func New(addr string) *Server {
	return &Server{
		Address: addr,
	}
}

func (h *Server) Serve(ctx context.Context, wg *sync.WaitGroup, errc chan<- error) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ai := audio.NewAudioInput()
		audioc, err := ai.Open()
		if err != nil {
			errc <- err
			return
		}
		defer ai.Close()

		bcast := audio.NewBroadcaster()
		aserver := audio.AudioServer{
			Broadcaster: bcast,
		}

		sfilter := audio.NewSilenceFilter()
		audioc = sfilter.Apply(audioc)
		bcast.Consume(audioc)

		srv := &http.Server{Addr: h.Address}

		http.HandleFunc("/audio", aserver.ServeAudio)
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "OK\n")
		})
		http.HandleFunc("/quitquitquit", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "Shutting down server in 1s\n")
			log.Warn("Stopping server in 1s due to /quitquitquit")
			go func() {
				time.Sleep(time.Second)
				srv.Shutdown(ctx)
			}()
		})

		go func() {
			<-ctx.Done()
			log.Info("Killing http server due to context cancel")
			srv.Shutdown(ctx)
		}()

		log.Infof("HTTP server listening on %q", h.Address)
		err = srv.ListenAndServe()

		if ctx.Err() != nil {
			log.Warnf("HTTP server terminated: %v", err)
		} else {
			log.Errorf("Http serve error: %v", err)
			// Report error since we're not already shutting down.
			errc <- err
		}
	}()
}
