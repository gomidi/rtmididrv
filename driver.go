package rtmididrv

import (
	"fmt"
	"strings"
	"sync"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/rtmididrv/imported/rtmidi"
)

type driver struct {
	opened []mid.Port
	sync.RWMutex
}

func (d *driver) String() string {
	return "rtmididrv"
}

// Close closes all open ports. It must be called at the end of a session.
func (d *driver) Close() (err error) {
	d.Lock()
	var e CloseErrors

	for _, p := range d.opened {
		err = p.Close()
		if err != nil {
			e = append(e, err)
		}
	}

	d.Unlock()

	if len(e) == 0 {
		return nil
	}

	return e
}

// New returns a driver based on the default rtmidi in and out
func New() (mid.Driver, error) {
	return &driver{}, nil
}

// Ins returns the available MIDI input ports
func (d *driver) Ins() (ins []mid.In, err error) {
	var in rtmidi.MIDIIn
	in, err = rtmidi.NewMIDIInDefault()
	if err != nil {
		return nil, fmt.Errorf("can't open default MIDI in: %v", err)
	}

	ports, err := in.PortCount()
	if err != nil {
		return nil, fmt.Errorf("can't get number of in ports: %s", err.Error())
	}

	for i := 0; i < ports; i++ {
		name, err := in.PortName(i)
		if err != nil {
			name = ""
		}
		ins = append(ins, newIn(d, i, name))
	}

	// don't destroy, destroy just panics
	// in.Destroy()
	err = in.Close()
	return
}

// Outs returns the available MIDI output ports
func (d *driver) Outs() (outs []mid.Out, err error) {
	var out rtmidi.MIDIOut
	out, err = rtmidi.NewMIDIOutDefault()
	if err != nil {
		return nil, fmt.Errorf("can't open default MIDI out: %v", err)
	}

	ports, err := out.PortCount()
	if err != nil {
		return nil, fmt.Errorf("can't get number of out ports: %s", err.Error())
	}

	for i := 0; i < ports; i++ {
		name, err := out.PortName(i)
		if err != nil {
			name = ""
		}
		outs = append(outs, newOut(d, i, name))
	}

	err = out.Close()
	return
}

type CloseErrors []error

func (c CloseErrors) Error() string {
	if len(c) == 0 {
		return "no errors"
	}

	var bd strings.Builder

	bd.WriteString("the following closing errors occured:\n")

	for _, e := range c {
		bd.WriteString(e.Error() + "\n")
	}

	return bd.String()
}
