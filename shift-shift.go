// Keyboard switcher for multiple keyboards.
// Left shift — group1
// Right shift - group2
package main

// go build -compiler gccgo -gccgoflags "-lX11" emacskey.go

// #cgo LDFLAGS: -lX11
// #include <stdlib.h>
// #include <stdio.h>
// #include <err.h>
// #include <X11/Xlib.h>
// #include <X11/XKBlib.h>
import "C"

import (
	"flag"
	"fmt"
	evdev "github.com/gvalkov/golang-evdev"
	"os"
	"os/signal"
	"strings"
	"time"
)

// Объединение данных для удобства передачи по каналу.
type Message struct {
	Device *evdev.InputDevice
	Events []evdev.InputEvent
}

func main() {
	var listDevices = flag.Bool("list", false, "list all devices listened by evdev")
	var printMode = flag.Bool("print", false, "print pressed keys")
	var quietMode = flag.Bool("quiet", false, "be silent")
	flag.Parse()

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)

	if *listDevices {
		scanDevices()
	} else {
		go listenKeyboards(*printMode, *quietMode)
		<-terminate
	}
}

// Переключалка групп Xorg.
func switchXkbGroup(group uint) {
	var xkbEventType, xkbError, xkbReason C.int
	var majorVers, minorVers C.int

	majorVers = C.XkbMajorVersion
	minorVers = C.XkbMinorVersion
	display := C.XkbOpenDisplay(nil, &xkbEventType, &xkbError, &majorVers, &minorVers, &xkbReason)
	if display == nil {
		fmt.Printf("Can't open X display %s", C.GoString(C.XDisplayName(nil)))
		switch xkbReason {
		case C.XkbOD_BadServerVersion:
		case C.XkbOD_BadLibraryVersion:
			println("incompatible versions of client and server XKB libraries")
		case C.XkbOD_ConnectionRefused:
			println("connection to X server refused")
		case C.XkbOD_NonXkbServer:
			println("XKB extension is not present")
		default:
			println("unknown error from XkbOpenDisplay: %d", xkbReason)
		}
		return
	}

	C.XkbLockGroup(display, C.XkbUseCoreKbd, C.uint(group))
	C.XCloseDisplay(display)
}

// Обнаруживает устройства, похожие на клавиатуры.
func scanDevices() (keyboards []*evdev.InputDevice) {
	for {
		devnames, err := evdev.ListInputDevices("/dev/input/event*")
		if err == nil && len(devnames) > 0 {
			for _, input := range devnames {
				dev, err := evdev.Open(input)
				if err != nil {
					fmt.Printf("Can't open %s\n", input)
					continue
				}
				if strings.Contains(strings.ToLower(dev.Name), "keyboard") {
					fmt.Printf("Keyboard found as %s [%s]\n", dev.Name, input)
					keyboards = append(keyboards, dev)
				}
			}
			return
		}
		time.Sleep(3 * time.Second)
	}
}

// Принимает события ото всех клавиатур.
func listenKeyboards(printMode, quietMode bool) {
	var lshift, rshift, lshift_started, rshift_started bool

scan:
	keyboards := scanDevices()
	if keyboards == nil {
		time.Sleep(5 * time.Second)
		goto scan
	}
	inbox := make(chan Message, 8)

	for _, kbd := range keyboards {
		if kbd == nil {
			continue
		}
		go listenEvents(kbd, inbox)
	}

	for {
		time.Sleep(5 * time.Millisecond)

		select {
		case msg := <-inbox:
			for _, ev := range msg.Events {
				switch ev.Type {
				case evdev.EV_SYN:
					// группа получена
					if (lshift && !rshift) || (rshift && !lshift) {
						lshift_started = false
						rshift_started = false
					} else {
						continue
					}
					if lshift {
						switchXkbGroup(C.XkbGroup1Index)
						if !quietMode {
							fmt.Printf("Left Shift at %s: English\n", msg.Device.Name)
						}
					} else if rshift {
						switchXkbGroup(C.XkbGroup2Index)
						if !quietMode {
							fmt.Printf("Right Shift at %s: Русский\n", msg.Device.Name)
						}
					}
					lshift = false
					rshift = false
				case evdev.EV_KEY:
					// обработка нажатий
					if printMode {
						fmt.Printf("%s: %v %v %d\n", msg.Device.Name, ev.Type, ev.Code, ev.Value)
					}
					switch ev.Value {
					case 1: // key down
						switch ev.Code {
						case evdev.KEY_LEFTSHIFT:
							lshift_started = true
						case evdev.KEY_RIGHTSHIFT:
							rshift_started = true
						default: // other keys
							lshift_started = false
							rshift_started = false
						}
					case 0: // key up
						switch ev.Code {
						case evdev.KEY_LEFTSHIFT:
							if lshift_started && !rshift_started {
								lshift = true
							}
						case evdev.KEY_RIGHTSHIFT:
							if rshift_started && !lshift_started {
								rshift = true
							}
						default:
							lshift_started = false
							rshift_started = false
						}
					}
				}
			}
		}
	}
}

// Слушает события и возвращает их в поток управления.
func listenEvents(kbd *evdev.InputDevice, replyTo chan Message) {
	for {
		events, err := kbd.Read()
		if err != nil || len(events) == 0 {
			//keyboards := scanDevices()
			//break
			// TODO
		}
		replyTo <- Message{Device: kbd, Events: events}
	}
}
