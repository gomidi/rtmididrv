package rtmididrv

import (
	"fmt"
	"math"
	"sync"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/rtmididrv/imported/rtmidi"
	//	"github.com/metakeule/mutex"
)

type in struct {
	driver *driver
	number int
	name   string
	midiIn rtmidi.MIDIIn
	sync.RWMutex
	//	mutex.RWMutex
	listenerSet bool
	closed      bool
}

// IsOpen returns wether the MIDI in port is open
func (i *in) IsOpen() (open bool) {
	i.RLock()
	open = !i.closed && i.midiIn != nil
	i.RUnlock()
	return
}

// String returns the name of the MIDI in port.
func (i *in) String() string {
	return i.name
}

// Underlying returns the underlying rtmidi.MIDIIn. Use it with type casting:
//   rtIn := i.Underlying().(rtmidi.MIDIIn)
func (i *in) Underlying() interface{} {
	return i.midiIn
}

// Number returns the number of the MIDI in port.
// Note that with rtmidi, out and in ports are counted separately.
// That means there might exists out ports and an in ports that share the same number.
func (i *in) Number() int {
	return i.number
}

// Close closes the MIDI in port, after it has stopped listening.
func (i *in) Close() error {
	i.RLock()
	if i.closed || i.midiIn == nil {
		i.RUnlock()
		return nil
	}
	i.RUnlock()

	i.Lock()
	i.closed = true
	i.stopListening()
	i.Unlock()

	//time.Sleep(time.Millisecond * 500)
	//i.Lock()
	err := i.midiIn.Close()
	//i.Unlock()
	if err != nil {
		return fmt.Errorf("can't close MIDI in port %v (%s): %v", i.number, i, err)
	}

	return nil
}

// Open opens the MIDI in port
func (i *in) Open() (err error) {
	i.RLock()
	if i.closed || i.midiIn != nil {
		i.RUnlock()
		return nil
	}
	i.RUnlock()

	i.Lock()
	defer i.Unlock()

	i.midiIn, err = rtmidi.NewMIDIInDefault()
	if err != nil {
		i.midiIn = nil
		return fmt.Errorf("can't open default MIDI in: %v", err)
	}

	err = i.midiIn.OpenPort(i.number, "")
	if err != nil {
		//i.midiIn.Destroy()
		i.midiIn = nil
		return fmt.Errorf("can't open MIDI in port %v (%s): %v", i.number, i, err)
	}

	i.driver.Lock()
	i.driver.opened = append(i.driver.opened, i)
	i.driver.Unlock()

	return nil
}

func newIn(debug bool, driver *driver, number int, name string) mid.In {
	i := &in{driver: driver, number: number, name: name}
	//	i.RWMutex = mutex.NewRWMutex("rtmididrv in port "+name, debug)
	return i
}

// SetListener makes the listener listen to the in port
func (i *in) SetListener(listener func(data []byte, deltaMicroseconds int64)) (err error) {
	i.RLock()
	if i.closed || i.midiIn == nil {
		i.RUnlock()
		return mid.ErrClosed
	}

	if i.listenerSet {
		i.RUnlock()
		return fmt.Errorf("listener allread set")
	}
	i.RUnlock()
	i.Lock()
	i.listenerSet = true
	i.Unlock()

	// since i.midiIn.SetCallback is blocking on success, there is no meaningful way to get an error
	// and set the callback non blocking
	go i.midiIn.SetCallback(func(_ rtmidi.MIDIIn, bt []byte, deltaSeconds float64) {
		// we want deltaMicroseconds as int64
		listener(bt, int64(math.Round(deltaSeconds*1000000)))
	})

	/*
		if err != nil {
			fmt.Errorf("can't set listener for MIDI in port %v (%s): %v", i.number, i, err)
		}
	*/
	return nil
}

// StopListening cancels the listening
func (i *in) StopListening() error {
	i.RLock()
	if i.closed || i.midiIn == nil {
		i.RUnlock()
		return mid.ErrClosed
	}
	i.RUnlock()
	i.Lock()
	err := i.stopListening()
	i.Unlock()
	return err
}

func (i *in) stopListening() error {
	err := i.midiIn.CancelCallback()
	if err != nil {
		fmt.Errorf("can't stop listening on MIDI in port %v (%s): %v", i.number, i, err)
	}
	return nil
}
