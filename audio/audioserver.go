package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{} // use default options
)

type AudioServer struct {
	Broadcaster *Broadcaster
}

func (a *AudioServer) ServeAudio(w http.ResponseWriter, r *http.Request) {
  fmt.Printf("Client connected with header\n%+v\n", r.Header)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("error upgrading websocket: ", err)
		return
	}
	defer c.Close()

	ar := a.Broadcaster.NewReceiver()
	defer ar.Close()

	// TODO close handler; some way to be responsive to connection closes.

	for data := range ar.GetAudioStream() {
		// TODO move this encoding nonsense before the broadcast.
		samples := []int16(data)

		buf := new(bytes.Buffer)
		buf.Grow(len(samples) * 2)

		// Encodes as PCM_16BIT
		for _, s := range samples {
			if err := binary.Write(buf, binary.LittleEndian, s); err != nil {
				fmt.Println("Binary write failure: ", err)
				return
			}
		}

    bs := buf.Bytes()

		if err := c.WriteMessage(websocket.BinaryMessage, bs); err != nil {
			fmt.Println("Audio websocket write failure: ", err)
			return
		}
	}
}
