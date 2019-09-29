package audio

import (
	"context"
	"fmt"
	"github.com/gordonklaus/portaudio"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

type AudioInput struct {
	ChunkSize   int
	DeviceName  string
	SampleRate  float64
	Broadcaster *Broadcaster
}

func NewAudioInput() *AudioInput {
	return &AudioInput{
		ChunkSize:  2048,
		DeviceName: "USB Audio Device",
		SampleRate: 48000,
	}
}

func (ai *AudioInput) OpenAndServe(ctx context.Context, wg *sync.WaitGroup) error {
	err := portaudio.Initialize()
	if err != nil {
		return err
	}

	log.Infof("AudioInput initialized %q", portaudio.VersionText())

	devices, err := portaudio.Devices()
	if err != nil {
		return err
	}
	var device *portaudio.DeviceInfo
	for _, d := range devices {
		if strings.Contains(d.Name, ai.DeviceName) {
			device = d
		}
	}
	if device == nil {
		return fmt.Errorf("Target not found in list of devices")
	}

	log.Infof("Device found! %q\n", device.Name)

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: 1,
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      ai.SampleRate,
		FramesPerBuffer: ai.ChunkSize,
	}

	buf := make([]int16, ai.ChunkSize)

	stream, err := portaudio.OpenStream(params, buf)
	if err != nil {
		return err
	}
	if err := stream.Start(); err != nil {
		return err
	}

	sf := NewSilenceFilter()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for ctx.Err() == nil {
			if err := stream.Read(); err != nil {
				log.Errorf("Portaudio read error: %v", err)
				// TODO make better, increment error counts, something nice.
				continue
			}

			samples := buf[:]

			sf.Process(samples)
			if sf.IsSilent() {
				continue
			}

			ai.Broadcaster.Broadcast(samples)
		}

		ai.Close()
	}()
	return nil
}

func (ai *AudioInput) Close() {
	if err := portaudio.Terminate(); err != nil {
		log.Errorf("Portaudio terminate failed: %v", err)
		return
	}
	log.Infof("Portaudio terminated")
}
