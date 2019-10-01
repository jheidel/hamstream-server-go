package audio

type AudioData []int16

const (
	ChunkSize  = 2048
	DeviceName = "USB Audio Device"
	SampleRate = 48000
)

// TODO move to audio.c
func (samples AudioData) GetLevel() int16 {
	max := int16(0)
	for _, s := range []int16(samples) {
		// SIGH, why doesn't go have absolute value builtin...
		if s > max {
			max = s
		}
		if -1*s > max {
			max = -1 * s
		}
	}
	return max
}
