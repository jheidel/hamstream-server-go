package audio

import (
	"bytes"
	"os/exec"

	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const (
	ESpeakBin = "/usr/bin/espeak"
	NiceBin   = "/usr/bin/nice"
)

func Speak(txt string) (*goaudio.IntBuffer, error) {
	cmd := exec.Command(NiceBin, "-n10", ESpeakBin, "--stdout", txt)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	w := wav.NewDecoder(bytes.NewReader(out))

	abuf, err := w.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	// Some crude resampling...
	// Duplicate samples to go from 22050 Hz to 44100 Hz which is closer
	// to our target 48000, so it won't sound so high pitched.
	d := make([]int, len(abuf.Data)*2)
	for i := 0; i < len(d); i++ {
		d[i] = abuf.Data[i/2]
	}
	abuf.Data = d
	abuf.Format.SampleRate *= 2

	return abuf, nil
}
