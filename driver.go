package rtmididrv

import (
	"fmt"
	"sync"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/rtmididrv/imported/rtmidi"
	//	"github.com/metakeule/mutex"
)

type driver struct {
	debug  bool
	opened []mid.Port
	sync.RWMutex
	//	mutex.RWMutex
	closed bool
}

func (d *driver) String() string {
	return "rtmididrv"
}

// Close closes all open ports. It must be called at the end of a session.
func (d *driver) Close() (err error) {

	d.RLock()
	if d.closed {
		d.RUnlock()
		return mid.ErrClosed
	}

	d.RUnlock()
	d.Lock()
	d.closed = true
	d.Unlock()

	for _, p := range d.opened {
		err = p.Close()
		// don't destroy, this just panics
		/*
			u := p.Underlying()
			switch v := u.(type) {
			case rtmidi.MIDIIn:
				v.Destroy()
			case rtmidi.MIDIOut:
				v.Destroy()
			}
		*/
	}

	// return just the last error to allow closing the other ports.
	// to ensure that all ports have been closed, this function must
	// return nil anyways
	return
}

// New returns a driver based on the default rtmidi in and out
//func New(debug bool) (connect.Driver, error) {
func New() (mid.Driver, error) {
	//d := &driver{debug: debug}
	d := &driver{}
	//	d.RWMutex = mutex.NewRWMutex("rtmididrv driver", debug)
	return d, nil
}

// Ins returns the available MIDI input ports
func (d *driver) Ins() (ins []mid.In, err error) {
	d.Lock()
	defer d.Unlock()

	if d.closed {
		return nil, mid.ErrClosed
	}
	in, err := rtmidi.NewMIDIInDefault()
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
		ins = append(ins, newIn(d.debug, d, i, name))
	}

	//in.Destroy()
	return
}

// Outs returns the available MIDI output ports
func (d *driver) Outs() (outs []mid.Out, err error) {
	d.Lock()
	defer d.Unlock()

	if d.closed {
		return nil, mid.ErrClosed
	}
	out, err := rtmidi.NewMIDIOutDefault()
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
		outs = append(outs, newOut(d.debug, d, i, name))
	}
	//out.Destroy()
	return
}
