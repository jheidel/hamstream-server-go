package audio

import(
  "fmt"
  "strings"
  "time"
  "github.com/gordonklaus/portaudio"
)

const (
  ChunkSize = 1024
  DeviceName = "USB Audio Device"
)

func streamCallback(in []int16) {
  fmt.Printf("Callback buffer: %v samples %v\n", len(in), in[:10])
}

func Test() error {
  fmt.Println("This is audio test.")

  err := portaudio.Initialize()
  if err != nil {
    return err
  }

  defer func() {
    err := portaudio.Terminate()
    if err != nil {
      fmt.Printf("Terminate failed: %v\n", err)
    }
    fmt.Println("Portaudio terminated")
  }()

  fmt.Println("Portaudio initialized: ", portaudio.VersionText())

  devices, err := portaudio.Devices()
  if err != nil {
    return err
  }
  var device *portaudio.DeviceInfo
  for _, d := range devices {
    if strings.Contains(d.Name, DeviceName) {
      device = d
    }
  }
  if device == nil {
    return fmt.Errorf("Target not found in list of devices")
  }

  fmt.Printf("Device found! %+v\n", device)

  params := portaudio.StreamParameters{
    Input: portaudio.StreamDeviceParameters{
      Device: device,
      Channels: 1,
      Latency: device.DefaultLowInputLatency,
    },
    SampleRate: 48000,
    FramesPerBuffer: ChunkSize,
  }

  stream, err := portaudio.OpenStream(params, streamCallback)
  if err != nil {
    fmt.Printf("Open stream error: %v\n", err)
    return err
  }

  err = stream.Start()
  if err != nil {
    fmt.Printf("Start stream error: %v\n", err)
    return nil
  }

  defer func() {
    err := stream.Close()
    if err != nil {
      fmt.Printf("Error while closing stream: %v\n", err)
    }
  }()



  time.Sleep(time.Second * 2)



  return nil
}
