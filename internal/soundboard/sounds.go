package soundboard

import (
	"os"
	"strings"

	"github.com/cwse1/jmidi/internal/logs"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/wav"
)

// Sound is a struct representing a sound to be played by the soundboard.
type Sound struct {
	path, name string
	format     beep.Format
}

// DecodeSound decodes sound file into a beep.Buffer
//
// Accepts a parameter 'sr' of type beep.SampleRate, which is the sample rate of the JACK server,
// to resample the sound if needed.
func (s *Sound) DecodeSound(sr beep.SampleRate) *beep.Buffer {
	f, err := os.Open(s.path) // Open sound file
	if err != nil {
		logs.Warn(err)
		return nil
	}
	defer f.Close()

	var decoder beep.StreamSeekCloser

	// Decode sound based on filetype
	if strings.HasSuffix(s.path, ".mp3") {
		decoder, s.format, err = mp3.Decode(f)
		if err != nil {
			logs.Warn(err)
			return nil
		}
	} else if strings.HasSuffix(s.path, ".wav") {
		decoder, s.format, err = wav.Decode(f)
		if err != nil {
			logs.Warn(err)
			return nil
		}
	}
	defer decoder.Close()

	resampler := beep.Resample(4, s.format.SampleRate, sr, decoder) // Resample audio file to match JACK server
	buffer := beep.NewBuffer(beep.Format{
		SampleRate:  sr,
		NumChannels: s.format.NumChannels,
		Precision:   s.format.Precision,
	})
	buffer.Append(resampler) // Add resampled audio to the returned buffer

	return buffer
}
