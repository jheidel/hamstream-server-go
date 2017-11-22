package audio

import (
	"sync"
	"time"
)

// SilenceFilter blocks an audio stream on input level.
type SilenceFilter struct {
	// LevelThresh defines the value at which samples are consisdered silent.
	LevelThresh int16
	// DurationThresh defines the time the silence must persist for before stopping the stream.
	DurationThresh time.Duration

	lastSound time.Time
	level     int16
	silent    bool
	mutex     sync.Mutex
}

func NewSilenceFilter() *SilenceFilter {
	return &SilenceFilter{
		LevelThresh:    50,
		DurationThresh: 5 * time.Second,
		silent:         true,
	}
}

// TODO move to audio.c
func getLevel(samples AudioData) int16 {
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

func (sf *SilenceFilter) Apply(in <-chan AudioData) <-chan AudioData {
	c := make(chan AudioData)
	go func() {
		defer close(c)
		for samples := range in {
			level := getLevel(samples)
			now := time.Now()
			if level > sf.LevelThresh {
				sf.lastSound = now
			}
			silent := true
			if now.Sub(sf.lastSound) < sf.DurationThresh {
				c <- samples
				silent = false
			}
			sf.mutex.Lock()
			sf.level = level
			sf.silent = silent
			sf.mutex.Unlock()
		}
	}()
	return c
}

func (sf *SilenceFilter) GetLevel() int16 {
	sf.mutex.Lock()
	defer sf.mutex.Unlock()
	return sf.level
}

func (sf *SilenceFilter) IsSilent() bool {
	sf.mutex.Lock()
	defer sf.mutex.Unlock()
	return sf.silent
}
