package audio

import (
	"context"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	StatsPeriod = time.Second / 15
)

type StatsServer struct {
	Broadcaster *Broadcaster
	Filter      *SilenceFilter
	Input       *AudioInput
}

type stats struct {
	Level       float32 `json:"level"`
	Gain        float32 `json:"gain"`
	AudioErrors int     `json:"aerrors"`
	Clients     int     `json:"clients"`
}

func (ss *StatsServer) Serve(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	log.Infof("Stats client %q connected with header %q", r.RemoteAddr, r.Header)
	defer func() {
		log.Infof("Stats client %q disconnected", r.RemoteAddr)
	}()

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("error upgrading websocket: %v", err)
		return
	}
	defer c.Close()

	pt := time.NewTicker(PingPeriod)
	t := time.NewTicker(StatsPeriod)
	for {
		select {
		case <-t.C:
			// Send json stats.
			s := &stats{
				Level:       float32(ss.Filter.GetLevel()) / (1<<15 - 1),
				Clients:     ss.Broadcaster.ReceiverCount(),
				AudioErrors: ss.Input.GetErrors(),
				Gain:        1.0,
				// TODO: gain
			}
			if err := c.WriteJSON(s); err != nil {
				log.Warnf("Stats write failure %v", err)
				return
			}

		case <-ctx.Done():
			return

		case now := <-pt.C:
			if err := c.WriteControl(websocket.PingMessage, []byte{}, now.Add(PingDeadline)); err != nil {
				log.Errorf("Failed to send stats ping: %v", err)
				return
			}
		}
	}
}
