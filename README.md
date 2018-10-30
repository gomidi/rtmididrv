# rtmididrv

[![Build Status Travis/Linux](https://travis-ci.org/gomidi/rtmididrv.svg?branch=master)](http://travis-ci.org/gomidi/rtmididrv)

## Purpose

A driver for the unified MIDI driver interface https://gitlab.com/gomidi/midi/mid.Driver .

This driver is based on the rtmidi project (see https://github.com/thestk/rtmidi for more information).
For a driver based on portmidi, see https://gitlab.com/gomidi/portmididrv

## Installation

It is recommended to use Go 1.11 with module support (`$GO111MODULE=on`).

## Linux / Debian

```
// install the headers of alsa somehow, e.g. sudo apt-get install libasound2-dev
go get -d gitlab.com/gomidi/rtmididrv
```

## Documentation

[![rtmididrv docs](http://godoc.org/gitlab.com/gomidi/rtmididrv?status.png)](http://godoc.org/gitlab.com/gomidi/rtmididrv)


## Example

```go
package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.com/gomidi/midi/mid"
	driver "gitlab.com/gomidi/rtmididrv"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// This example expects the first input and output port to be connected
// somehow (are either virtual MIDI through ports or physically connected).
// We write to the out port and listen to the in port.
func main() {
	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	if len(os.Args) == 2 && os.Args[1] == "list" {
		printInPorts(ins)
		printOutPorts(outs)
		return
	}

	in, out := ins[0], outs[0]

	must(in.Open())
	must(out.Open())

	wr := mid.WriteTo(out)

	// listen for MIDI
	go mid.NewReader().ReadFrom(in)

	{ // write MIDI to out that passes it to in on which we listen.
		err := wr.NoteOn(60, 100)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Nanosecond)
		wr.NoteOff(60)
		time.Sleep(time.Nanosecond)

		wr.SetChannel(1)

		wr.NoteOn(70, 100)
		time.Sleep(time.Nanosecond)
		wr.NoteOff(70)
		time.Sleep(time.Second * 1)
	}
}

func printPort(port mid.Port) {
	fmt.Printf("[%v] %s\n", port.Number(), port.String())
}

func printInPorts(ports []mid.In) {
	fmt.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}

func printOutPorts(ports []mid.Out) {
	fmt.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}

```
