package audio

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	log "github.com/sirupsen/logrus"
	"strings"
)

type AudioInput struct {
	ChunkSize  int
	DeviceName string
	SampleRate float64

	audioc chan AudioData
	quit   chan chan error
}

func NewAudioInput() *AudioInput {
	return &AudioInput{
		ChunkSize:  2048,
		DeviceName: "USB Audio Device",
		SampleRate: 48000,
		quit:       make(chan chan error),
	}
}

func (ai *AudioInput) Open() (<-chan AudioData, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	log.Infof("AudioInput initialized %q", portaudio.VersionText())

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}
	var device *portaudio.DeviceInfo
	for _, d := range devices {
		if strings.Contains(d.Name, ai.DeviceName) {
			device = d
		}
	}
	if device == nil {
		return nil, fmt.Errorf("Target not found in list of devices")
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

	ai.audioc = make(chan AudioData)

	//callback := func(in []int16) {
	//	ai.audioc <- AudioData(in)
	//}

	//stream, err := portaudio.OpenStream(params, callback)

	stream, err := portaudio.OpenStream(params, buf)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	// TODO... in tight loop
	// 1) Read from audio source
	// 2) Do the signal processing
	// 3) Copy to any registered output buffers.

	go func() {
		for {
			// TODO: don't malloc here. Do something better with the chaining here.
			n := make([]int16, ai.ChunkSize)
			if err := stream.Read(); err != nil {
				log.Errorf("Portaudio read error: %v", err)
				// TODO make better, increment error counts, something nice.
				continue
			}
			copy(n, buf)
			ai.audioc <- n

			select {
			case c := <-ai.quit:
				err := stream.Close()
				c <- err
				return
			default:
			}
		}
	}()

	return ai.audioc, nil
}

func (ai *AudioInput) Close() {
	c := make(chan error)
	ai.quit <- c
	err := <-c
	if err != nil {
		log.Errorf("Error while closing stream: %v\n", err)
	}

	err = portaudio.Terminate()
	if err != nil {
		fmt.Printf("Terminate failed: %v\n", err)
	}
	fmt.Println("Portaudio terminated")
}
