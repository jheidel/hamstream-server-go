package audio

import (
	"fmt"
)

type Broadcaster struct {
	// addc adds new channels to the broadcast set
	addc chan *Receiver

	// removec removes channels from the broadcast set
	removec chan *Receiver

	// receivers contains the set of channels to broadcast data to
	receivers map[int]*Receiver

	// genID is used to produce monotonically increasing receiver IDs.
	genID int
}

type Receiver struct {
	ID     int
	audioc chan AudioData

	b *Broadcaster
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		addc:      make(chan *Receiver),
		removec:   make(chan *Receiver),
		receivers: make(map[int]*Receiver),
	}
}

func (r *Receiver) GetAudioStream() <-chan AudioData {
	return r.audioc
}

func (r *Receiver) Close() {
	r.b.removec <- r
}

func (b *Broadcaster) NewReceiver() *Receiver {
	r := &Receiver{
		ID:     b.genID,
		audioc: make(chan AudioData, 100),
		b:      b,
	}
	b.genID += 1

	// NOTE: this will panic if the broadcaster has shut down.
	b.addc <- r
	return r
}

func (b *Broadcaster) Consume(audioc <-chan AudioData) {
	go func() {
		closed := false
		for {
			select {
			case r := <-b.addc:
				if closed {
					close(r.audioc)
				} else {
					b.receivers[r.ID] = r
					fmt.Println("Receiver added ", r.ID)
				}
			case r := <-b.removec:
				delete(b.receivers, r.ID)
				fmt.Println("Receiver removed ", r.ID)
				if !closed {
					close(r.audioc)
				}
			case data, ok := <-audioc:
				if !ok {
					for _, r := range b.receivers {
						close(r.audioc)
					}
					closed = true
				} else {
					for _, r := range b.receivers {
						r.audioc <- data
					}
				}
			}
			if closed && len(b.receivers) == 0 {
				close(b.addc)
				fmt.Println("Clean exit of broadcast")
				return
			}
		}
	}()
}
