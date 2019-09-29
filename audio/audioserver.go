package audio

import (
	"context"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type AudioServer struct {
	Broadcaster *Broadcaster
}

func (a *AudioServer) Serve(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	log.Infof("Client %q connected with header %q", r.RemoteAddr, r.Header)
	defer func() {
		log.Infof("Client %q disconnected", r.RemoteAddr)
	}()

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("error upgrading websocket: %v", err)
		return
	}
	defer c.Close()

	ar := a.Broadcaster.NewReceiver()
	defer ar.Close()

	pt := time.NewTicker(PingPeriod)

	audioc := ar.GetAudioStream(ctx)
	for {
		select {
		case data, ok := <-audioc:
			if !ok {
				log.Infof("Disconnecting %q, end of stream", r.RemoteAddr)
				return
			}

			if err := c.WriteMessage(websocket.BinaryMessage, data); err != nil {
				log.Warnf("Write failure %v", err)
				return
			}

		case <-ctx.Done():
			return

		case now := <-pt.C:
			if err := c.WriteControl(websocket.PingMessage, []byte{}, now.Add(PingDeadline)); err != nil {
				log.Errorf("Failed to send ping: %v", err)
				return
			}
		}
	}
}
