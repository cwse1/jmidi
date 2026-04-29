# jmidi

Client for controlling JACK with MIDI devices. Based on [jack_mixer]("https://github.com/jack-mixer/jack_mixer").

## Features

- jack_mixer controls using MIDI notes instead of CC messages
- Soundboard functionality

### Planned Features

- Custom commands
- Sampling
- Granular lighting control
- Layers
- automatic config reloads

## Usage

Run jmidi after your JACK server starts and connect its ports where necessary.
Each client will have a MIDI input port, where the MIDI controller for that
client can be connected.
