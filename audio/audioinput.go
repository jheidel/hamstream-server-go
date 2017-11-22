package audio

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"strings"
)

type AudioInput struct {
	ChunkSize  int
	DeviceName string
	SampleRate float64

	stream *portaudio.Stream
	audioc chan AudioData
}

func NewAudioInput() *AudioInput {
	return &AudioInput{
		ChunkSize:  1024,
		DeviceName: "USB Audio Device",
		SampleRate: 48000,
	}
}

func (ai *AudioInput) Open() (<-chan AudioData, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	fmt.Println("AudioInput initialized: ", portaudio.VersionText())

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

	fmt.Printf("Device found! %+v\n", device)

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: 1,
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      ai.SampleRate,
		FramesPerBuffer: ai.ChunkSize,
	}

	ai.audioc = make(chan AudioData)

	callback := func(in []int16) {
		ai.audioc <- AudioData(in)
	}

	stream, err := portaudio.OpenStream(params, callback)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	ai.stream = stream
	return ai.audioc, nil
}

func (ai *AudioInput) Close() {
	err := ai.stream.Close()
	if err != nil {
		fmt.Printf("Error while closing stream: %v\n", err)
	}

	err = portaudio.Terminate()
	if err != nil {
		fmt.Printf("Terminate failed: %v\n", err)
	}
	fmt.Println("Portaudio terminated")
}
