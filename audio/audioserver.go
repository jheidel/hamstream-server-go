package audio

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{} // use default options
)

type AudioServer struct {
	Broadcaster *Broadcaster
}

func (a *AudioServer) ServeAudio(w http.ResponseWriter, r *http.Request) {
	log.Infof("Client %q connected with header %q", r.RemoteAddr, r.Header)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("error upgrading websocket: %v", err)
		return
	}
	defer c.Close()

	ar := a.Broadcaster.NewReceiver()
	defer ar.Close()

	audioc := ar.GetAudioStream()
	for {
		select {
		case data, ok := <-audioc:
			if !ok {
				log.Infof("Disconnecting %q, end of stream", r.RemoteAddr)
				return
			}

			if err := c.WriteMessage(websocket.BinaryMessage, data); err != nil {
				log.Warnf("Client %q disconnect, write failure %v", r.RemoteAddr, err)
				return
			}

		case <-r.Context().Done():
			log.Infof("Client %q disconnected", r.RemoteAddr)
			return
		}
	}
}
