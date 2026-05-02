// Package jackclient implements a JACK client using the go-jack library.
package jackclient

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cwse1/jmidi/internal/config"
	"github.com/cwse1/jmidi/internal/logs"
	"github.com/cwse1/jmidi/internal/soundboard"

	"github.com/gopxl/beep"
	"github.com/xthexder/go-jack"
)

var (
	controller   *config.Controller
	library      *soundboard.Library
	lightsOffMsg jack.MidiData
)

// Client is a JACK client with midi ports, an output port, state maps, beep.Mixer, and a pointer to its process callback.
type Client struct {
	jackClient                              *jack.Client
	midiInPort, midiReturnPort, midiOutPort *jack.Port
	outPorts                                []*jack.Port
	state                                   map[byte]byte
	ccMap                                   map[string]byte
	processPtr                              func(uint32) int
	wg                                      sync.WaitGroup
	isStopping, isRunning                   bool
	sr                                      beep.SampleRate
	Mixer                                   *beep.Mixer
}

// Create new client
func newClient() *Client {
	client := &Client{
		state:      make(map[byte]byte),
		ccMap:      make(map[string]byte),
		isRunning:  true,
		isStopping: false,
		Mixer:      &beep.Mixer{},
	}

	// Points to the active process callback for this client
	client.processPtr = client.startProcess

	var code int
	client.jackClient, code = jack.ClientOpen("jmidi-"+controller.Name, jack.NoStartServer) // Open JACK client and connect to server
	if client.jackClient == nil {
		logs.Fatal(fmt.Errorf("JACK :- Could not create a client and connect to server; exited with code: %d", code))
	}

	// Register client ports
	client.midiInPort = client.jackClient.PortRegister("midi_in", jack.DEFAULT_MIDI_TYPE, jack.PortIsInput, 0)
	client.midiReturnPort = client.jackClient.PortRegister("midi_ret", jack.DEFAULT_MIDI_TYPE, jack.PortIsOutput, 0)
	client.midiOutPort = client.jackClient.PortRegister("midi_out", jack.DEFAULT_MIDI_TYPE, jack.PortIsOutput, 0)
	client.outPorts = append(client.outPorts, client.jackClient.PortRegister("out_L", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0))
	client.outPorts = append(client.outPorts, client.jackClient.PortRegister("out_R", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0))

	// Set JACK client callbacks
	client.jackClient.OnShutdown(func() {
		client.stop()
	})
	if code := client.jackClient.SetPortConnectCallback(func(_, _ jack.PortId, _ bool) {
		client.isRunning = false
	}); code != 0 {
		logs.Fatal(fmt.Errorf("JACK :- SetPortConnectCallback exited with code: %d", code))
	}
	if code := client.jackClient.SetProcessCallback(func(nframes uint32) int {
		return client.processPtr(nframes)
	}); code != 0 {
		logs.Fatal(fmt.Errorf("JACK :- SetProcessCallback exited with code: %d", code))
	}

	client.sr = beep.SampleRate(client.jackClient.GetSampleRate()) // Get sample rate of JACK server

	return client
}

// Initialize client state
func (client *Client) initState() *Client {
	// Turn on lights for sound pads
	for key := range controller.SoundboardPads {
		client.state[key] = 1
	}

	// Read jack_mixer config, return client if err
	conf, err := config.ReadJMConfig()
	if err != nil {
		logs.Warn(err)
		return client
	}

	// Initialize lighting states
	for _, ic := range conf.InChannel {
		client.ccMap[ic.Name] = ic.MuteCC
		for key, pad := range controller.ControlPads {
			if pad.Target == ic.Name {
				if ic.Muted {
					client.state[key] = 1
				} else {
					client.state[key] = 0
				}
			}
		}
	}
	for _, oc := range conf.OutChannel {
		client.ccMap[oc.Name] = oc.MuteCC
		for key, pad := range controller.ControlPads {
			if pad.Target == oc.Name {
				if oc.Muted {
					client.state[key] = 1
				} else {
					client.state[key] = 0
				}
			}
		}
	}

	for key := range client.state {
		lightsOffMsg.Buffer = append(lightsOffMsg.Buffer, 144, key, 0)
	}

	return client
}

// Start client processing
func (client *Client) start() {
	if code := client.jackClient.Activate(); code != 0 {
		logs.Fatal(fmt.Errorf("JACK :- Failed to activate client; exited with code: %d", code))
	}
}

// Callback function run when program or JACK server are shut
func (client *Client) stop() {
	client.processPtr = client.stopProcess
	client.isStopping = true
	client.wg.Add(2)

	client.wg.Wait()
	client.jackClient.Close()
}

// TODO: implement support for more states
// Toggle lighting state
func (client *Client) lightState(state byte) jack.MidiData {
	midiData := jack.MidiData{}
	for key := range client.state {
		if client.state[key] != 0 {
			midiData.Buffer = append(midiData.Buffer, 144, key, state)
		}
	}
	return midiData
}

// Run Client passing associated controller and the global audio library.
func Run(ctrl *config.Controller, lib *soundboard.Library) {
	controller = ctrl
	library = lib

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	client := newClient().initState()
	client.start()

	if lib.SampleRate != client.sr {
		lib.SampleRate = client.sr // Set sample rate of audio library
	}

	<-sig // Catch SIGTERM or SIGINT and gracefully handle shutting down
	client.stop()
	close(sig)
}
