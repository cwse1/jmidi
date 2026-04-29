package jackclient

import (
	"github.com/xthexder/go-jack"
)

// Process callback run at startup; sets initial lighting state
func (client *Client) startProcess(nframes uint32) int {
	if !client.isRunning {
		ret := client.lightState(1) // Midi message to turn all relevant lights on

		bufMidiRet := client.midiReturnPort.MidiClearBuffer(nframes)
		client.midiReturnPort.MidiEventWrite(&ret, bufMidiRet)

		client.isRunning = true                // Set client state to running
		client.processPtr = client.mainProcess // Point to the main process callback
	}
	return 0
}

// Process callback run at shutdown; turns off all lights
func (client *Client) stopProcess(nframes uint32) int {
	if client.isStopping {
		ret := client.lightState(0)

		bufMidiRet := client.midiReturnPort.MidiClearBuffer(nframes)
		client.midiReturnPort.MidiEventWrite(&ret, bufMidiRet)

		client.wg.Done()
	}
	return 0
}

// Main process callback; reads midi input data; returns lighting midi data; can send midi data to jack_mixer
func (client *Client) mainProcess(nframes uint32) int {
	out := jack.MidiData{}
	ret := jack.MidiData{}

	bufMidiIn := client.midiInPort.GetMidiEvents(nframes)
	bufMidiRet := client.midiReturnPort.MidiClearBuffer(nframes)
	bufMidiOut := client.midiOutPort.MidiClearBuffer(nframes)
	bufOutL := client.outPorts[0].GetBuffer(nframes)
	bufOutR := client.outPorts[1].GetBuffer(nframes)

	// loop over midi input
	for _, sample := range bufMidiIn {
		key := sample.Buffer[1] // note value of pad
		switch sample.Buffer[0] {
		case 144:
			if pad, ok := controller.ControlPads[key]; ok {
				switch pad.Function {
				case "mute": // mute jack_mixer channel
					switch client.state[key] {
					case 0:
						client.state[key] = 1
						out.Buffer, ret.Buffer = []byte{176, client.ccMap[pad.Target], 127}, []byte{144, key, 1}
					case 1:
						client.state[key] = 0
						out.Buffer, ret.Buffer = []byte{176, client.ccMap[pad.Target], 0}, []byte{144, key, 0}
					default:
					}
					client.midiOutPort.MidiEventWrite(&out, bufMidiOut)
					client.midiReturnPort.MidiEventWrite(&ret, bufMidiRet)
				case "toggleLights": // toggles all lights
					switch client.state[key] {
					case 0:
						client.state[key] = 1
						ret = client.lightState(0)
					case 1:
						client.state[key] = 0
						ret = client.lightState(1)
					default:
					}
					client.midiReturnPort.MidiEventWrite(&ret, bufMidiRet)
				}
			}
			if pad, ok := controller.SoundboardPads[key]; ok {
				switch pad.Function {
				case "": // queues sound from audio library
					if library != nil {
						if s := library.PlaySound(pad.Target, pad.Volume); s != nil {
							client.Mixer.Add(s)
						}
					}
				case "stop": // stops all sounds
					client.Mixer.Clear()
				}
			}
		case 128:
		default:
		}
	}
	// process sound queue
	frames := make([][2]float64, nframes)
	client.Mixer.Stream(frames)
	for i := range min(len(frames), int(nframes)) {
		bufOutL[i] = jack.AudioSample(frames[i][0])
		bufOutR[i] = jack.AudioSample(frames[i][1])
	}
	return 0
}
