package main

import (
	"github.com/cwse1/jmidi/internal/config"
	"github.com/cwse1/jmidi/internal/jackclient"
	"github.com/cwse1/jmidi/internal/soundboard"
)

func main() {
	// TODO: implement live config reloading, and a command line interface

	conf := config.Read() // Read config file

	lib := soundboard.NewAudioLibrary(conf.SoundsPath) // Initialize soundboard

	// Initialize JACK client for each controller
	for _, controller := range conf.Controllers {
		jackclient.Run(&controller, lib)
	}
}
