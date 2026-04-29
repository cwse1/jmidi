// Package soundboard provides soundboard functionality to JACK clients.
package soundboard

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cwse1/jmidi/internal/logs"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
)

// Library is the model used by the soundboard. It holds sound file locations and a decoded cache.
type Library struct {
	sounds     map[string]*Sound
	cache      map[string]*beep.Buffer
	cacheMu    sync.RWMutex
	SampleRate beep.SampleRate
}

// NewAudioLibrary initializes the audio library.
func NewAudioLibrary(path string) *Library {
	lib := &Library{
		sounds: make(map[string]*Sound),
		cache:  make(map[string]*beep.Buffer),
	}

	// get sound files from sound dir
	files, err := os.ReadDir(path)
	if err != nil {
		logs.Warn(err)
		return nil
	}

	// add sounds to library
	for _, file := range files {
		// trim file suffix
		basename := strings.TrimSuffix(file.Name(), ".mp3")
		basename = strings.TrimSuffix(basename, ".wav")

		lib.sounds[basename] = &Sound{
			path: filepath.Join(path, file.Name()),
			name: basename,
		}
	}
	return lib
}

// PlaySound returns sound streamer by reading from cache or decoding sound into cache
// A volume modifier is applied if defined.
func (lib *Library) PlaySound(target string, volume float64) *effects.Volume {
	lib.cacheMu.Lock()
	defer lib.cacheMu.Unlock()

	if _, ok := lib.cache[target]; !ok {
		if lib.cache[target] = lib.sounds[target].DecodeSound(lib.SampleRate); lib.cache[target] == nil {
			return nil
		}
	}

	return &effects.Volume{
		Streamer: lib.cache[target].Streamer(0, lib.cache[target].Len()),
		Base:     2,
		Volume:   volume,
		Silent:   false,
	}
}
