package audio

import (
	"context"
	"os"
	"sync"

	"container/ring"
	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"
	log "github.com/sirupsen/logrus"
)

const (
	WavRingSize = 256
)

type WavWriter struct {
	Path string

	buf  *ring.Ring
	bufc chan []int
}

func (w *WavWriter) OpenAndHost(ctx context.Context, wg *sync.WaitGroup) error {
	w.buf = ring.New(WavRingSize)
	w.bufc = make(chan []int, WavRingSize-1)
	for i := 0; i < WavRingSize; i++ {
		w.buf.Value = make([]int, ChunkSize)
		w.buf = w.buf.Next()
	}

	// TODO dynamic filename from date.
	fn := w.Path
	wf, err := os.Create(fn)
	if err != nil {
		return err
	}

	we := wav.NewEncoder(wf, 48000, 16, 1, 1)
	log.Infof("Opened WAV %q", fn)

	// TODO dynamic
	header, err := Speak("TEST TEST TEST this is a radio stream recording TEST TEST TEST")
	if err != nil {
		return err
	}
	if err := we.Write(header); err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			log.Infof("Closing WAV %q", fn)
			if err := we.Close(); err != nil {
				log.Errorf("Failed to close WAV encoder: %v", err)
			}
			if err := wf.Close(); err != nil {
				log.Errorf("Failed to close WAV file: %v", err)
			}
		}()

		for ctx.Err() == nil {
			select {
			case buf := <-w.bufc:
				// New sample!
				abuf := &goaudio.IntBuffer{
					Format: &goaudio.Format{
						NumChannels: 1,
						SampleRate:  48000,
					},
					Data:           buf,
					SourceBitDepth: 16,
				}
				if err := we.Write(abuf); err != nil {
					log.Errorf("Failed to write sample to WAV: %v", err)
				}

			case <-ctx.Done():
				return
			}
		}

	}()
	return nil
}

func (w *WavWriter) Write(samples AudioData) {
	buf := w.buf.Value.([]int)[:]
	w.buf = w.buf.Next()
	for i := 0; i < len(samples); i++ {
		buf[i] = int(samples[i])
	}
	select {
	case w.bufc <- buf[:]:
	default:
		log.Error("Dropped wav data, would block!")
	}
}
