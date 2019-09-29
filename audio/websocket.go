package audio

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	PingPeriod   = 10 * time.Second
	PingDeadline = 10 * time.Second
)

var (
	upgrader = websocket.Upgrader{} // use default options
)
