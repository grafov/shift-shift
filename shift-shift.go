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
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gvalkov/golang-evdev"
)

// Объединение данных для удобства передачи по каналу.
type Message struct {
	Device *evdev.InputDevice
	Events []evdev.InputEvent
}

func main() {
	var (
		listDevices  = flag.Bool("list", false, "list all devices listened by evdev")
		printMode    = flag.Bool("print", false, "print pressed keys")
		quietMode    = flag.Bool("quiet", false, "be silent")
		deviceMatch  = flag.String("match", "keyboard", "string used to match keyboard device")
		keysymFirst  = flag.String("first", "LEFTSHIFT", "key used for switcing on first xkb group")
		keysymSecond = flag.String("second", "RIGHTSHIFT", "key used for switcing on second xkb group")
	)

	flag.Parse()

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)

	if *listDevices {
		for devicePath, device := range getInputDevices() {
			fmt.Printf("%s %s\n", devicePath, device.Name)
		}
	} else {
		keyFirst, err := getKeyCode(*keysymFirst)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		keySecond, err := getKeyCode(*keysymSecond)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		go listenKeyboards(keyFirst, keySecond, *printMode, *quietMode, *deviceMatch)

		<-terminate
	}
}

// Возвращает keycode по текстому представлению клавиши.
func getKeyCode(keysym string) (uint16, error) {
	for evdevKeyCode, evdevKeySym := range evdev.KEY {
		if evdevKeySym == "KEY_"+keysym {
			return uint16(evdevKeyCode), nil
		}
	}

	return 0, fmt.Errorf("keycode for keysym %s not found", keysym)
}

// Переключалка групп Xorg.
func switchXkbGroup(group uint) {
	var xkbEventType, xkbError, xkbReason C.int
	var majorVers, minorVers C.int

	majorVers = C.XkbMajorVersion
	minorVers = C.XkbMinorVersion
	display := C.XkbOpenDisplay(nil, &xkbEventType, &xkbError, &majorVers, &minorVers, &xkbReason)
	if display == nil {
		log.Printf("unable to open X display %s", C.GoString(C.XDisplayName(nil)))

		switch xkbReason {
		case C.XkbOD_BadServerVersion:
		case C.XkbOD_BadLibraryVersion:
			log.Printf("incompatible versions of client and server XKB libraries")
		case C.XkbOD_ConnectionRefused:
			log.Printf("connection to X server refused")
		case C.XkbOD_NonXkbServer:
			log.Printf("XKB extension is not present")
		default:
			log.Printf("unknown error from XkbOpenDisplay: %d", xkbReason)
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
				log.Printf("unable to open device %s: %s", devicePath, err)

				continue
			}

			inputDevices[devicePath] = device
		}
	}

	return inputDevices
}

// Обнаруживает устройства, похожие на клавиатуры.
func scanDevices(mbox chan Message, deviceMatch string, quietMode bool) {
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
						if !quietMode {
							log.Printf(
								"new keyboard at %s: %s",
								devicePath, device.Name,
							)
						}

						if mbox != nil {
							keyboards[devicePath] = device

							go listenEvents(
								devicePath, device, mbox, kbdLost, quietMode,
							)
						}
					}
				}
			}

			time.Sleep(4 * time.Second)
		}
	}
}

// Принимает события ото всех клавиатур.
func listenKeyboards(
	keyFirst uint16, keySecond uint16,
	printMode, quietMode bool, deviceMatch string,
) {
	var groupFirst, groupSecond, groupFirstEnabled, groupSecondEnabled bool

	inbox := make(chan Message, 8)
	kbdLost := make(chan bool, 8)
	kbdLost <- true // init

	go scanDevices(inbox, deviceMatch, quietMode)

	for {
		select {
		case msg := <-inbox:
			for _, ev := range msg.Events {
				switch ev.Type {
				case evdev.EV_SYN:
					// группа получена
					if (groupFirst && !groupSecond) || (groupSecond && !groupFirst) {
						groupFirstEnabled = false
						groupSecondEnabled = false
					} else {
						continue
					}

					if groupFirst {
						switchXkbGroup(C.XkbGroup1Index)
						if !quietMode {
							log.Printf("switching to first group at %s", msg.Device.Name)
						}
					} else if groupSecond {
						switchXkbGroup(C.XkbGroup2Index)
						if !quietMode {
							log.Printf("switching to second group at %s", msg.Device.Name)
						}
					}
					groupFirst = false
					groupSecond = false
				case evdev.EV_KEY:
					// обработка нажатий
					if printMode {
						log.Printf(
							"%s: %v %v %d",
							msg.Device.Name, ev.Type, ev.Code, ev.Value,
						)
					}

					switch ev.Value {
					case 1: // key down
						switch ev.Code {
						case keyFirst:
							groupFirstEnabled = true
						case keySecond:
							groupSecondEnabled = true
						default: // other keys
							groupFirstEnabled = false
							groupSecondEnabled = false
						}
					case 0: // key up
						switch ev.Code {
						case keyFirst:
							if groupFirstEnabled && !groupSecondEnabled {
								groupFirst = true
							}
						case keySecond:
							if groupSecondEnabled && !groupFirstEnabled {
								groupSecond = true
							}
						default:
							groupFirstEnabled = false
							groupSecondEnabled = false
						}
					}
				}
			}
		}
	}
}

func listenEvents(
	name string,
	kbd *evdev.InputDevice,
	replyTo chan Message,
	kbdLost chan string,
	quietMode bool,
) {
	for {
		events, err := kbd.Read()
		if err != nil || len(events) == 0 {
			if !quietMode {
				log.Printf("keyboard %s disconnected", kbd.Name)
			}

			kbdLost <- name
			return
		}

		replyTo <- Message{Device: kbd, Events: events}
	}
}
