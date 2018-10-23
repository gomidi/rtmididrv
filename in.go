package rtmididrv

import (
	"fmt"
	"math"

	"github.com/gomidi/connect"
	"github.com/gomidi/rtmididrv/imported/rtmidi"
	"github.com/metakeule/mutex"
)

type in struct {
	driver *driver
	number int
	name   string
	midiIn rtmidi.MIDIIn
	mutex.RWMutex
}

// IsOpen returns wether the MIDI in port is open
func (i *in) IsOpen() (open bool) {
	i.RLock()
	open = i.midiIn != nil
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
	//	fmt.Println("rtmididrv close called")
	i.RLock()
	//	fmt.Println("rtmididrv close read lock acquired")
	if i.midiIn == nil {
		i.RUnlock()
		return nil
	}
	i.RUnlock()
	//	fmt.Println("rtmididrv close read lock released")

	i.Lock()
	//	fmt.Println("rtmididrv close lock acquired")
	defer i.Unlock()

	i.stopListening()
	//	fmt.Println("rtmididrv close stopListening done")

	err := i.midiIn.Close()
	//	fmt.Println("rtmididrv close inner close called")
	if err != nil {
		return fmt.Errorf("can't close MIDI in port %v (%s): %v", i.number, i, err)
	}

	i.midiIn = nil

	//i.midiIn.Destroy()
	return nil
}

// Open opens the MIDI in port
func (i *in) Open() (err error) {
	i.RLock()
	if i.midiIn != nil {
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

func newIn(debug bool, driver *driver, number int, name string) connect.In {
	i := &in{driver: driver, number: number, name: name}
	i.RWMutex = mutex.NewRWMutex("rtmididrv in port "+name, debug)
	return i
}

// SetListener makes the listener listen to the in port
func (i *in) SetListener(listener func(data []byte, deltaMicroseconds int64)) error {
	i.RLock()
	if i.midiIn == nil {
		i.RUnlock()
		return connect.ErrClosed
	}
	i.Lock()
	defer i.Unlock()
	err := i.midiIn.SetCallback(func(_ rtmidi.MIDIIn, bt []byte, deltaSeconds float64) {
		// we want deltaMicroseconds as int64
		listener(bt, int64(math.Round(deltaSeconds*1000000)))
	})
	if err != nil {
		fmt.Errorf("can't set listener for MIDI in port %v (%s): %v", i.number, i, err)
	}
	return nil
}

// StopListening cancels the listening
func (i *in) StopListening() error {
	//	panic("stop listening called")
	//	fmt.Println("stop listening called")
	i.RLock()
	//	fmt.Println("stop listening rlock acquired")
	if i.midiIn == nil {
		i.RUnlock()
		return connect.ErrClosed
	}
	i.RUnlock()
	//	fmt.Println("stop listening rlock released")
	i.Lock()
	//	fmt.Println("stop listening lock acquired")
	defer func() {
		i.Unlock()
		//		fmt.Println("stop listening lock released")
	}()
	return i.stopListening()
}

func (i *in) stopListening() error {
	//	fmt.Println("stoplistening")
	err := i.midiIn.CancelCallback()
	if err != nil {
		fmt.Errorf("can't stop listening on MIDI in port %v (%s): %v", i.number, i, err)
	}
	//	fmt.Println("done stoplistening")
	return nil
}
