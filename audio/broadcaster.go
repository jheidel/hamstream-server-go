package audio

import (
	"bytes"
	"container/ring"
	"context"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"sync"
)

const (
	// TODO compute these based on the audio sample rates.
	RingCount = 128  // About 5s of audio
	BufSize   = 4096 // based on chunk size of 2048 of int16s
)

type Broadcaster struct {
	// receivers contains the set of channels to broadcast data to
	receivers map[int]*Receiver

	// genID is used to produce monotonically increasing receiver IDs.
	genID int

	// protects receivers state.
	mu sync.RWMutex

	// backs PCM conversion
	pcmBuf *bytes.Buffer
}

type Receiver struct {
	ID     int
	parent *Broadcaster

	head *ring.Ring
	tail *ring.Ring

	closed bool
	mu     sync.Mutex
	cond   *sync.Cond
}

func (r *Receiver) Broadcast(data []byte) {
	buf := r.head.Value.(*bytes.Buffer)
	buf.Reset()
	if _, err := buf.Write(data); err != nil {
		panic(err)
	}

	r.mu.Lock()
	r.head = r.head.Next()
	overrun := r.head == r.tail
	if overrun {
		// Need to move the tail out of the way....
		// TODO: but this could introduce data corruption...
		r.tail = r.tail.Next()
	}
	r.cond.Signal()
	r.mu.Unlock()

	if overrun {
		log.Warn("Receiver ring buffer overrun, data lost!")
	}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		receivers: make(map[int]*Receiver),
		pcmBuf:    new(bytes.Buffer),
	}
}

func (b *Broadcaster) Broadcast(samples AudioData) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.receivers) == 0 {
		return
	}

	// Convert sample to PCM.
	b.pcmBuf.Reset()
	b.pcmBuf.Grow(len(samples) * 2)
	for _, s := range samples {
		if err := binary.Write(b.pcmBuf, binary.LittleEndian, s); err != nil {
			panic(err)
		}
	}

	// Write PCM data to receivers.
	for _, r := range b.receivers {
		r.Broadcast(b.pcmBuf.Bytes())
	}
}

func (b *Broadcaster) ReceiverCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.receivers)
}

func (b *Broadcaster) NewReceiver() *Receiver {
	// Initialize the ring buffer.
	rbuf := ring.New(RingCount)
	for i := 0; i < RingCount; i++ {
		b := new(bytes.Buffer)
		b.Grow(BufSize)
		rbuf.Value = b
		rbuf = rbuf.Next()
	}

	r := &Receiver{
		parent: b,

		head: rbuf,
		tail: rbuf,
	}
	r.cond = sync.NewCond(&r.mu)

	b.mu.Lock()
	r.ID = b.genID
	b.genID += 1
	b.receivers[r.ID] = r
	b.mu.Unlock()

	log.Infof("Receiver %d added", r.ID)
	return r
}

func (r *Receiver) Close() {
	log.Infof("Receiver %d removed", r.ID)
	r.parent.mu.Lock()
	delete(r.parent.receivers, r.ID)
	r.parent.mu.Unlock()
}

func (r *Receiver) GetAudioStream(ctx context.Context) <-chan []byte {
	c := make(chan []byte)
	go func() {
		<-ctx.Done()
		r.mu.Lock()
		defer r.mu.Unlock()
		r.closed = true
		r.cond.Signal()
	}()
	go func() {
		defer close(c)
		for {
			// TODO what happens when we run into our tail....

			r.mu.Lock()
			for r.head == r.tail && !r.closed {
				// Wait for new data.
				r.cond.Wait()
			}
			if r.closed {
				return
			}
			buf := r.tail.Value.(*bytes.Buffer)
			r.tail = r.tail.Next()
			r.mu.Unlock()

			c <- buf.Bytes()
		}
	}()
	return c
}
