// Keyboard layout switcher for Xorg/XKB and Wayland/Sway.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/grafov/shift-shift/hyprland"
	"github.com/grafov/shift-shift/sway"
	"github.com/grafov/shift-shift/xkb"

	"github.com/grafov/evdev"
)

const (
	scanPeriod            = 15 * time.Second
	scanPeriodNoKeyboards = 2 * time.Second
)

type switcher interface {
	Init() error
	Switch(int)
	Name() string
	Close()
}

// Объединение данных для удобства передачи по каналу.
type message struct {
	Device *evdev.InputDevice
	Events []evdev.InputEvent
}

func main() {
	var err error
	listDevices := flag.Bool("list", false, `list all devices that found by evdev (not only keyboards)`)
	listSwayDevices := flag.Bool("list-sway", false, `list all devices recognized by Sway (not only keyboards)`)
	listHyprlandDevices := flag.Bool("list-hypr", false, `list all keyboards recognized by Hyprland`)
	printMode := flag.Bool("print", false, `print pressed keys for debug (verbose output)`)
	quietMode := flag.Bool("quiet", false, `be silent`)
	kbdRegex := flag.String("match", "keyboard", `regexp used to match input keyboard device as it listed by evdev`)
	wmRegex := flag.String("match-wm", "keyboard", `optional regexp used to match device in WM, if not set evdev regexp used`)
	keysym1 := flag.String("1", "LEFTSHIFT", `key used for switching to 1st xkb group`)
	keysym2 := flag.String("2", "RIGHTSHIFT", `key used for switching to 2nd xkb group`)
	keysym3 := flag.String("3", "", `key used for switching to 3rd xkb group`)
	keysym4 := flag.String("4", "", `key used for switching to 4th xkb group`)
	scanOnce := flag.Bool("scan-once", false, `scan for keyboards only at startup (less power consumption)`)
	dblKeystroke := flag.Bool("double-keystroke", false, `require pressing the same key twice to switch the layout`)
	dblKeyTimeout := flag.Int("double-keystroke-timeout", 500, `second keystroke timeout in milliseconds`)
	switchMethod := flag.String("switcher", "auto", `select method of switching (possible values are "auto", "xkb", "sway", "hypr")`)

	flag.Parse()
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)

	if *listDevices {
		for devicePath, device := range getInputDevices() {
			fmt.Printf("%s %s\n", devicePath, device.Name)
		}
		return
	}
	if *listSwayDevices {
		var list []string
		if list, err = sway.PrintDevices(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for _, inp := range list {
			fmt.Println(inp)
		}
		return
	}

	if *listHyprlandDevices {
		var list []string
		if list, err = hyprland.PrintDevices(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for _, inp := range list {
			fmt.Println(inp)
		}
		return
	}

	switches := make([]uint16, 4)
	for i, k := range []string{*keysym1, *keysym2, *keysym3, *keysym4} {
		if k == "" {
			continue
		}
		switches[i], err = getKeyCode(k)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	matchEvdevKbds, err := regexp.Compile(*kbdRegex)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"unable to compile regexp for matching devices of evdev: %s\n",
			err,
		)
		os.Exit(1)
	}

	if wmRegex == nil {
		wmRegex = kbdRegex
	}

	matchWMKbds, err := regexp.Compile(*wmRegex)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"unable to compile regexp for matching devices in WM: %s\n",
			err,
		)
		os.Exit(1)
	}

	var sw switcher
	switch *switchMethod {
	case "hypr":
		if sw, err = hyprland.New(matchWMKbds, scanPeriod, scanPeriodNoKeyboards, *scanOnce, *printMode); err != nil {
			break
		}
	case "sway":
		sw = sway.New(matchWMKbds, scanPeriod, scanPeriodNoKeyboards, *scanOnce, *printMode)
	case "xkb":
		sw = xkb.New()
	case "auto":
		fallthrough
	default:
		switch {
		case hyprland.CheckAvailability():
			if sw, err = hyprland.New(matchWMKbds, scanPeriod, scanPeriodNoKeyboards, *scanOnce, *printMode); err != nil {
				break
			}
		case sway.CheckAvailability():
			sw = sway.New(matchWMKbds, scanPeriod, scanPeriodNoKeyboards, *scanOnce, *printMode)
		default:
			sw = xkb.New()
		}
	}
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"unable to start switcher: %s\n",
			err,
		)
		os.Exit(1)
	}

	if !*quietMode {
		log.Printf("use %s switcher\n", sw.Name())
	}
	if err = sw.Init(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"unable to init switcher: %s\n",
			err,
		)
		os.Exit(1)
	}

	go listenKeyboards(sw,
		switches,
		*printMode, *quietMode,
		matchEvdevKbds, *scanOnce,
		*dblKeystroke, *dblKeyTimeout,
	)

	<-terminate
	sw.Close()
}

// Возвращает keycode по текстому представлению клавиши.
func getKeyCode(keysym string) (uint16, error) {
	for evdevKeyCode, evdevKeySym := range evdev.KEY {
		if evdevKeySym == "KEY_"+keysym {
			return uint16(evdevKeyCode), nil
		}
	}

	return 0, fmt.Errorf("keycode for keysym %s is not found", keysym)
}

func getInputDevices() map[string]*evdev.InputDevice {
	inputDevices := make(map[string]*evdev.InputDevice)

	devicePaths, err := evdev.ListInputDevicePaths("/dev/input/event*")
	if err == nil && len(devicePaths) > 0 {
		for _, devicePath := range devicePaths {
			device, err := evdev.Open(devicePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to open device %s: %s", devicePath, err)
				continue
			}
			inputDevices[devicePath] = device
		}
	}
	return inputDevices
}

// Обнаруживает устройства, похожие на клавиатуры.
func scanDevices(mbox chan message, deviceMatch *regexp.Regexp, quietMode bool, scanOnce bool) {
	keyboards := make(map[string]*evdev.InputDevice)
	kbdLost := make(chan string, 8)
	var notified bool
	for {
		select {
		case input := <-kbdLost:
			delete(keyboards, input)
			notified = false
		default:
			for devicePath, device := range getInputDevices() {
				if deviceMatch.MatchString(device.Name) {
					if _, ok := keyboards[devicePath]; !ok {
						if !quietMode {
							log.Printf("evdev keyboard found at %s: %s", devicePath, device.Name)
						}
						if mbox != nil {
							keyboards[devicePath] = device
							go listenEvents(devicePath, device, mbox, kbdLost, quietMode)
						}
					}
				} else {
					err := device.File.Close()
					if err != nil {
						log.Printf("unable to close device %q: %s", device.Name, err)
					}
				}
			}
			if scanOnce {
				return
			}
			if len(keyboards) == 0 {
				if !quietMode && !notified {
					log.Print("no keyboards connected")
					notified = true
				}
				time.Sleep(scanPeriodNoKeyboards)
			} else {
				time.Sleep(scanPeriod)
			}
		}
	}
}

// Принимает события ото всех клавиатур.
func listenKeyboards(
	sw switcher,
	switches []uint16,
	printMode, quietMode bool, deviceMatch *regexp.Regexp,
	scanOnce bool, dblKeystroke bool, dblKeyTimeout int,
) {
	var (
		useGroup int
		t        *time.Timer
	)

	inbox := make(chan message, 8)
	kbdLost := make(chan bool, 8)
	kbdLost <- true // init
	if dblKeystroke {
		t = time.NewTimer(time.Duration(dblKeyTimeout) * time.Millisecond)
	}

	go scanDevices(inbox, deviceMatch, quietMode, scanOnce)

	var prevKey evdev.InputEvent
	for msg := range inbox {
		for _, ev := range msg.Events {
			if ev.Type != evdev.EV_KEY {
				continue
			}
			if prevKey.Code == ev.Code && prevKey.Value == ev.Value && prevKey.Type == ev.Type {
				continue
			}
			prevKey.Type = ev.Type
			prevKey.Code = ev.Code
			prevKey.Value = ev.Value
			if printMode {
				var pv string
				switch ev.Value {
				case 0:
					pv = "released"
				case 1:
					pv = "pressed"
				case 2:
					pv = "hold"
				default:
					pv = "undefined status"
				}
				log.Printf("%s type:%v code:%v %s", msg.Device.Name, ev.Type, ev.Code, pv)
			}
			switch ev.Value {
			case 1: // key down
				useGroup = 0
				for i, k := range switches {
					if ev.Code != k {
						continue
					}
					ready := true
					if dblKeystroke {
						t, ready = checkTimeout(t, dblKeyTimeout)
					}
					if ready {
						useGroup = i + 1
					}
				}
			case 0: // key up
				if useGroup == 0 {
					break
				}
				for i, k := range switches {
					if ev.Code != k {
						continue
					}
					if useGroup == i+1 {
						if printMode {
							log.Printf("%s switches group to %d", msg.Device.Name, useGroup)
						}
						sw.Switch(useGroup)
					}
				}
				useGroup = 0
			}
		}
	}
}

// Check new key presses/releases from choosen keyboards.
func listenEvents(
	name string,
	kbd *evdev.InputDevice,
	replyTo chan message,
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

		replyTo <- message{Device: kbd, Events: events}
	}
}

// Checks if the key was pressed before the timer was expired
// Resets expired timer
func checkTimeout(t *time.Timer, timeout int) (*time.Timer, bool) {
	if t.Stop() {
		return t, true
	}
	t.Reset(time.Duration(timeout) * time.Millisecond)
	return t, false
}
