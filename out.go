package rtmididrv

import (
	"fmt"
	"sync"
	"time"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/rtmididrv/imported/rtmidi"
)

func newOut(debug bool, driver *driver, number int, name string) mid.Out {
	o := &out{driver: driver, number: number, name: name}
	return o
}

type out struct {
	driver  *driver
	midiOut rtmidi.MIDIOut
	number  int
	name    string
	sync.RWMutex
	closed bool
}

// IsOpen returns wether the port is open
func (o *out) IsOpen() (open bool) {
	o.RLock()
	open = !o.closed && o.midiOut != nil
	o.RUnlock()
	return
}

// Send sends a message to the MIDI out port
// If the out port is closed, it returns mid.ErrClosed
func (o *out) Send(b []byte) error {
	//o.RLock()
	o.Lock()
	defer o.Unlock()
	if o.closed || o.midiOut == nil {
		//o.RUnlock()
		return mid.ErrClosed
	}
	//	o.RUnlock()

	err := o.midiOut.SendMessage(b)
	if err != nil {
		return fmt.Errorf("could not send message to MIDI out %v (%s): %v", o.number, o, err)
	}
	return nil
}

// Underlying returns the underlying rtmidi.MIDIOut. Use it with type casting:
//   rtOut := o.Underlying().(rtmidi.MIDIOut)
func (o *out) Underlying() interface{} {
	return o.midiOut
}

// Number returns the number of the MIDI out port.
// Note that with rtmidi, out and in ports are counted separately.
// That means there might exists out ports and an in ports that share the same number
func (o *out) Number() int {
	return o.number
}

// String returns the name of the MIDI out port.
func (o *out) String() string {
	return o.name
}

// Close closes the MIDI out port
func (o *out) Close() error {
	o.RLock()
	if o.closed || o.midiOut == nil {
		o.RUnlock()
		return nil
	}
	o.RUnlock()
	o.Lock()
	o.closed = true
	o.Unlock()

	// disabling closing of the out port. it does not work reliably in a context with multiple goroutines
	// last try for closing
	time.Sleep(time.Millisecond * 500)
	//	o.Lock()
	err := o.midiOut.Close()
	//	o.midiOut.Destroy()
	//	o.Unlock()

	if err != nil {
		return fmt.Errorf("can't close MIDI out %v (%s): %v", o.number, o, err)
	}

	return nil
}

// Open opens the MIDI out port
func (o *out) Open() (err error) {
	o.RLock()
	if o.closed || o.midiOut != nil {
		o.RUnlock()
		return nil
	}
	o.RUnlock()
	o.Lock()
	defer o.Unlock()
	o.midiOut, err = rtmidi.NewMIDIOutDefault()
	if err != nil {
		return fmt.Errorf("can't open default MIDI out: %v", err)
	}

	err = o.midiOut.OpenPort(o.number, "")
	if err != nil {
		return fmt.Errorf("can't open MIDI out port %v (%s): %v", o.number, o, err)
	}

	o.driver.Lock()
	o.driver.opened = append(o.driver.opened, o)
	o.driver.Unlock()

	return nil
}
