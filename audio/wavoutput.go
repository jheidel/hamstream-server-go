package audio

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"container/ring"
	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"
	strftime "github.com/jehiah/go-strftime"
	log "github.com/sirupsen/logrus"
)

const (
	WavRingSize = 256

	// Generate a section header if this much time elapsed since the last sample.
	SectionThresh = 30 * time.Second
)

func genFileName() string {
	return fmt.Sprintf("%s.wav", strftime.Format("%Y%m%d-%H%M%S", time.Now()))
}

func genFileSpeech() string {
	return fmt.Sprintf("Radio stream recording from %s.", strftime.Format("%B %-d, %Y", time.Now()))
}

func genSectionSpeech() string {
	return strftime.Format("%-H %M", time.Now())
}

func speakToFile(txt string, we *wav.Encoder) error {
	d, err := Speak(txt)
	if err != nil {
		return err
	}
	if err := we.Write(d); err != nil {
		return err
	}
	return nil
}

type WavWriter struct {
	PathBase string

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

	if err := os.MkdirAll(w.PathBase, 0660); err != nil {
		return fmt.Errorf("Failed to create recording directory: %v", err)
	}
	fn := path.Join(w.PathBase, genFileName())
	wf, err := os.Create(fn)
	if err != nil {
		return err
	}

	we := wav.NewEncoder(wf, 48000, 16, 1, 1)
	log.Infof("Opened WAV %q", fn)

	if err := speakToFile(genFileSpeech(), we); err != nil {
		return fmt.Errorf("Failed to generate file speech header: %v", err)
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

		var lastSpeech time.Time

		for ctx.Err() == nil {
			select {
			case buf := <-w.bufc:
				// Write section header if enough time elapsed.
				now := time.Now()
				if now.Sub(lastSpeech) > SectionThresh {
					if err := speakToFile(genSectionSpeech(), we); err != nil {
						log.Errorf("Failed to generate section speech: %v", err)
					}
				}
				lastSpeech = now

				// Write new sample!
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
