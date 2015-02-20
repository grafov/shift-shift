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
	"os"

	"github.com/gvalkov/golang-evdev/evdev"
	// "os/exec"
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
	var (
		listDevices = flag.Bool("list", false, "list all devices listened by evdev")
		printMode   = flag.Bool("print", false, "print pressed keys")
		quietMode   = flag.Bool("quiet", false, "be silent")
		deviceMatch = flag.String("match", "keyboard", "string used to match keyboard device")
	)
	flag.Parse()

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)

	if *listDevices {
		for devicePath, device := range getInputDevices() {
			fmt.Printf("[%s] %s\n", devicePath, device.Name)
		}
	} else {
		go listenKeyboards(*printMode, *quietMode, *deviceMatch)
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

func getInputDevices() map[string]*evdev.InputDevice {
	inputDevices := make(map[string]*evdev.InputDevice)

	devicePaths, err := evdev.ListInputDevicePaths("/dev/input/event*")
	if err == nil && len(devicePaths) > 0 {
		for _, devicePath := range devicePaths {
			device, err := evdev.Open(devicePath)
			if err != nil {
				fmt.Printf("Can't open %s\n", devicePath)
				continue
			}

			inputDevices[devicePath] = device
		}
	}

	return inputDevices
}

// Обнаруживает устройства, похожие на клавиатуры.
func scanDevices(mbox chan Message, deviceMatch string) {
	var keyboards map[string]*evdev.InputDevice = make(map[string]*evdev.InputDevice)

	kbdLost := make(chan string, 8)

	for {
		select {
		case input := <-kbdLost:
			delete(keyboards, input)
		default:
			for devicePath, device := range getInputDevices() {
				if strings.Contains(strings.ToLower(device.Name), strings.ToLower(deviceMatch)) {
					if _, ok := keyboards[devicePath]; !ok {
						fmt.Printf("Keyboard found as %s [%s]\n", device.Name, devicePath)
						if mbox != nil {
							keyboards[devicePath] = device
							go listenEvents(devicePath, device, mbox, kbdLost)
							// TODO эти гвозди надо выковырять
							// exec.Command("sudo", "-u", "$USER", "setxkbmap", "-layout", "us+typo,ru:2+typo", "-option", "lv3:ralt_switch").Start()
							// exec.Command("sudo", "-u", "$USER", "xset", "r", "rate", "300", "55").Start()
						}
					}
				}
			}
			time.Sleep(4 * time.Second)
		}
	}
}

// Принимает события ото всех клавиатур.
func listenKeyboards(printMode, quietMode bool, deviceMatch string) {
	var lshift, rshift, lshift_started, rshift_started bool

	inbox := make(chan Message, 8)
	kbdLost := make(chan bool, 8)
	kbdLost <- true // init

	go scanDevices(inbox, deviceMatch)

	for {
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
func listenEvents(name string, kbd *evdev.InputDevice, replyTo chan Message, kbdLost chan string) {
	for {
		events, err := kbd.Read()
		if err != nil || len(events) == 0 {
			fmt.Printf("Keyboard %s lost...\n", kbd.Name)
			kbdLost <- name
			return
		}
		replyTo <- Message{Device: kbd, Events: events}
	}
}
