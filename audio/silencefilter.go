package audio

import (
	"sync"
	"time"
)

// TODO: make this a generic audio stats module.

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

func (sf *SilenceFilter) Process(sample AudioData) {
	level := sample.GetLevel()
	now := time.Now()
	if level > sf.LevelThresh {
		sf.lastSound = now
	}
	silent := true
	if now.Sub(sf.lastSound) < sf.DurationThresh {
		silent = false
	}
	sf.mutex.Lock()
	sf.level = level
	sf.silent = silent
	sf.mutex.Unlock()
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
