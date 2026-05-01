// Package config handles reading config files.
package config

// TODO: use Viper/Cobra

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cwse1/jmidi/internal/logs"
	"go.yaml.in/yaml/v4"
)

// Path to sounds directory; default: $XDG_CONFIG_HOME/jmidi/sounds/
var sounds string

// Config holds values read from the jmidi config file.
type Config struct {
	SoundsPath  string
	Controllers []Controller `yaml:"controllers"`
}

// jack_mixer channel config and mute state
type channel struct {
	Name   string `xml:"name,attr"`
	Muted  bool   `xml:"out_mute,attr"`
	MuteCC byte   `xml:"mute_midi_cc,attr"`
}

// Holds values read from the jack_mixer config file
type jackMixerConfig struct {
	Config     xml.Name  `xml:"jack_mixer"`
	InChannel  []channel `xml:"input_channel"`
	OutChannel []channel `xml:"output_channel"`
}

// Find user config directory
func getUserConfigDir() string {
	path, err := os.UserConfigDir()
	if err != nil {
		logs.Fatal(fmt.Errorf("failed to find system config dir: %e", err))
	}

	return path
}

// Create jmidi config file
func createConf(path string) {
	exampleConf := `
controllers: [
	{
		name: "controller-name",
		controlPads: {
		# midi_note: {function: "mute" | "toggleLights", target: "Jack-Mixer-Channel"}
			# when target is undefined function cannot be undefined
			0: {function: "mute", target: "Mic"}, # toggles the Mic channel in jack_mixer
			1: {function: "toggleLights"}, # toggles midi controller lighting
		},
		soundboardPads: {
		# midi_note: {function: "stop" | "", target: "sound-file-no-suffix" | "", volume: 1 | ""}
			# when function is undefined, target is defined and is a sound file without its suffix: eg. snare-drum.mp3 -> snare-drum
			# when target is undefined function cannot be undefined
			# volume can be defined on soundboard pads and will play the sound at v=2^volume; so a volume of -1 is 0.5=2^(-1)
			# volume being undefined defaults to 0 which will produce a sound at 1=2^0 which enacts no change
			2: {target: "snare-drum", volume: 1}, # plays a snare-drum sound at 200% volume
			3: {target: "symbal", volume: -1}, # plays a symbal sound at 50% volume
			4: {function: "stop"}, # stops all sounds playing
		},
	},
]
`

	err := os.MkdirAll(sounds, os.ModeDir) // Create $XDG_CONFIG_HOME/jmidi/sounds/ directory
	if err != nil {
		logs.Warn(err)
	}

	conf, err := os.Create(path) // Create config file
	if err != nil {
		logs.Fatal(fmt.Errorf("failed to create config file: %e", err))
	}

	_, err = conf.Write([]byte(exampleConf)) // Write example config to file
	if err != nil {
		logs.Fatal(fmt.Errorf("failed to write example config file: %e", err))
	}

	err = conf.Close() // Close config file
	if err != nil {
		logs.Fatal(fmt.Errorf("failed to close config file: %e", err))
	}
}

// getConf reads the config file into a Config struct and returns it
func getConf(path string) Config {
	var conf Config

	file, err := os.Open(path) // Open config file
	if err != nil {
		logs.Fatal(fmt.Errorf("failed to open config file: %e", err))
	}
	defer file.Close()

	data, err := io.ReadAll(file) // Read config file
	if err != nil {
		logs.Fatal(fmt.Errorf("failed to read config file: %e", err))
	}

	if err := yaml.Unmarshal(data, &conf); err != nil {
		logs.Fatal(fmt.Errorf("failed to unmarshal config file: %e", err))
	}

	conf.SoundsPath = sounds

	return conf
}

// Read checks for an exisiting config file, calls createConf if none is found, and returns the result of getConf
func Read() Config {
	path := getUserConfigDir()

	cfg := filepath.Join(path, "jmidi", "config.yaml")
	sounds = filepath.Join(filepath.Dir(cfg), "sounds")

	_, err := os.Stat(cfg) // Check for existing config file
	if err != nil {
		logs.Warn(fmt.Errorf("config file not found; creating: %e", err))

		createConf(cfg)
	}

	return getConf(cfg)
}

// ReadJMConfig reads the jack_mixer config file into a jackMixerConfig struct and returns it.
// Errors if not found or could not be read.
func ReadJMConfig() (jackMixerConfig, error) {
	path := getUserConfigDir()

	cfg, err := os.ReadFile(filepath.Join(path, "jack_mixer", "config.xml"))
	if err != nil {
		return jackMixerConfig{}, err
	}

	var conf jackMixerConfig

	err = xml.Unmarshal(cfg, &conf)
	if err != nil {
		return jackMixerConfig{}, err
	}
	return conf, nil
}
