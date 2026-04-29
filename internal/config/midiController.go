package config

// Controller is a struct that holds a configuration of a single midi controller.
type Controller struct {
	Name           string       `yaml:"name"`
	ControlPads    map[byte]Pad `yaml:"controlPads"`
	SoundboardPads map[byte]Pad `yaml:"soundboardPads"`
}

// Pad is a struct that represents a midi pad on a controller.
type Pad struct {
	Function string  `yaml:"function"`
	Target   string  `yaml:"target"`
	Volume   float64 `yaml:"volume"`
}
