package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/mid"
	driver "gitlab.com/gomidi/rtmididrv"
)

var (
	portsMx sync.Mutex
	drv     mid.Driver

	inPorts  = map[int]mid.In{}
	outPorts = map[int]mid.Out{}
)

func init() {
	var err error
	drv, err = driver.New()
	if err != nil {
		panic("can't initialize driver")
	}
}

func main() {
	// make sure to close all open ports at the end
	defer drv.Close()

	var ww = make(chan int, 10)

	go func() {
		for {
			go checkPorts()
			time.Sleep(time.Second * 1)
		}
	}()

	// interrupt with ctrl+c
	<-ww
}

func greet(out mid.Out) {
	out.Open()
	wr := mid.ConnectOut(out)
	time.Sleep(time.Millisecond * 200)
	wr.NoteOn(60, 100)
	time.Sleep(time.Nanosecond)
	wr.NoteOff(60)
	time.Sleep(time.Nanosecond)
	wr.SetChannel(1)
	wr.NoteOn(70, 100)
	time.Sleep(time.Nanosecond)
	wr.NoteOff(70)
	time.Sleep(time.Second * 1)
}

func listen(in mid.In) {
	in.Open()
	rd := mid.NewReader(mid.NoLogger())
	rd.Msg.Each = func(_ *mid.Position, msg midi.Message) {
		fmt.Printf("got message %s from in port %s\n", msg.String(), in.String())
	}
	mid.ConnectIn(in, rd)
}

func checkPorts() {
	//fmt.Println("...")
	portsMx.Lock()
	ins, _ := drv.Ins()

	for _, in := range ins {
		if strings.Contains(in.String(), "Client") {
			continue
		}
		if inPorts[in.Number()] != nil {
			if inPorts[in.Number()].String() != in.String() {
				inPorts[in.Number()].StopListening()
				inPorts[in.Number()].Close()
				fmt.Printf("closing in port: [%v] %s\n", in.Number(), inPorts[in.Number()].String())
				inPorts[in.Number()] = in
				fmt.Printf("new in port: [%v] %s\n", in.Number(), in.String())
				go listen(in)
			} else {
				continue
			}
		} else {
			inPorts[in.Number()] = in
			fmt.Printf("new in port: [%v] %s\n", in.Number(), in.String())
			go listen(in)
		}
	}

	outs, _ := drv.Outs()

	for _, out := range outs {
		if strings.Contains(out.String(), "Client") {
			continue
		}
		if outPorts[out.Number()] != nil {
			if outPorts[out.Number()].String() != out.String() {
				outPorts[out.Number()].Close()
				fmt.Printf("closing out port: [%v] %s\n", out.Number(), outPorts[out.Number()].String())
				outPorts[out.Number()] = out
				fmt.Printf("new out port: [%v] %s\n", out.Number(), out.String())
				go greet(out)
			} else {
				continue
			}
		} else {
			fmt.Printf("new out port: [%v] %s\n", out.Number(), out.String())
			outPorts[out.Number()] = out
			go greet(out)
		}
	}
	portsMx.Unlock()
}
