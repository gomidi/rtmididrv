package main

import (
	"fmt"
	"os"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/rtmididrv"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// run this in two terminals. first terminal without args to create the virtual ports and
// second terminal with argument "list" to see the ports.
func main() {
	drv, err := rtmididrv.New()
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

	var in mid.In
	in, err = drv.OpenVirtualIn("test-virtual-in")

	must(err)

	var out mid.Out
	out, err = drv.OpenVirtualOut("test-virtual-out")

	must(err)

	wr := mid.ConnectOut(out)

	// listen for MIDI
	rd := mid.NewReader()
	// example to write received messages from the virtual in port to the virtual out port
	rd.Msg.Each = func(_ *mid.Position, msg midi.Message) {
		wr.Write(msg)
	}
	c := make(chan int, 10)
	go mid.ConnectIn(in, rd)
	<-c
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
