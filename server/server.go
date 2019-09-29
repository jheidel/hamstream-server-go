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
	Address     string
	AudioServer *audio.AudioServer
	StatsServer *audio.StatsServer
}

func (h *Server) Serve(ctx context.Context, wg *sync.WaitGroup, errc chan<- error) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		srv := &http.Server{Addr: h.Address}

		http.HandleFunc("/audio", h.AudioServer.Serve)
		http.HandleFunc("/stats", h.StatsServer.Serve)
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
		err := srv.ListenAndServe()

		if ctx.Err() != nil {
			log.Warnf("HTTP server terminated: %v", err)
		} else {
			log.Errorf("Http serve error: %v", err)
			// Report error since we're not already shutting down.
			errc <- err
		}
	}()
}
