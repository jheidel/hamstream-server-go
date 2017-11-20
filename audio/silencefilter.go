package audio

import (
	"time"
)

// SilenceFilter blocks an audio stream on input level.
type SilenceFilter struct {
	// LevelThresh defines the value at which samples are consisdered silent.
	LevelThresh int16
	// DurationThresh defines the time the silence must persist for before stopping the stream.
	DurationThresh time.Duration

	lastSound time.Time
}

func NewSilenceFilter() *SilenceFilter {
	return &SilenceFilter{
		LevelThresh:    50,
		DurationThresh: 5 * time.Second,
	}
}

func getLevel(samples []int16) int16 {
	max := int16(0)
	for _, s := range samples {
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

func (sf *SilenceFilter) Apply(in <-chan []int16) <-chan []int16 {
	c := make(chan []int16)
	go func() {
    defer close(c)
		for samples := range in {
			level := getLevel(samples)
			now := time.Now()
			if level > sf.LevelThresh {
				sf.lastSound = now
			}
			if now.Sub(sf.lastSound) < sf.DurationThresh {
				c <- samples
			}
		}
	}()
	return c
}
