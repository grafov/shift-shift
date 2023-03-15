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

	"github.com/grafov/shift-shift/sway"
	"github.com/grafov/shift-shift/xkb"

	evdev "github.com/gvalkov/golang-evdev"
)

const scanPeriod = 4 * time.Second

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
	printMode := flag.Bool("print", false, `print pressed keys for debug (verbose output)`)
	quietMode := flag.Bool("quiet", false, `be silent`)
	kbdRegex := flag.String("match", "keyboard", `regexp used to match keyboard device`)
	keysym1 := flag.String("1", "LEFTSHIFT", `key used for switching to 1st xkb group`)
	keysym2 := flag.String("2", "RIGHTSHIFT", `key used for switching to 2nd xkb group`)
	keysym3 := flag.String("3", "", `key used for switching to 3rd xkb group`)
	keysym4 := flag.String("4", "", `key used for switching to 4th xkb group`)
	scanOnce := flag.Bool("scan-once", false, `scan for keyboards only at startup (less power consumption)`)
	dblKeystroke := flag.Bool("double-keystroke", false, `require pressing the same key twice to switch the layout`)
	dblKeyTimeout := flag.Int("double-keystroke-timeout", 500, `second keystroke timeout in milliseconds`)
	switchMethod := flag.String("switcher", "auto", `select method of switching (possible values are "auto", "xkb", "sway")`)

	flag.Parse()
	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)

	if *listDevices {
		for devicePath, device := range getInputDevices() {
			fmt.Printf("%s %s\n", devicePath, device.Name)
		}
		return
	}
	if *listSwayDevices {
		var list []sway.SwayInput
		if list, err = sway.GetInputDevices(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for _, inp := range list {
			fmt.Printf("%s type:%s name:%s layout:%s group:%d\n",
				inp.Identifier, inp.Type, inp.Name, inp.XkbActiveLayoutName, inp.XkbActiveLayoutIndex)
		}
		return
	}

	switches := make([]uint16, 4, 4)
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

	matchDevs, err := regexp.Compile(*kbdRegex)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"unable to compile regexp for matching devices: %s\n",
			err,
		)
		os.Exit(1)
	}

	var sw switcher
	switch *switchMethod {
	case "sway":
		sw = sway.New(matchDevs, scanPeriod, *scanOnce, *printMode)
	case "xkb":
		sw = xkb.New()
	case "auto":
		fallthrough
	default:
		if sway.CheckAvailability() {
			sw = sway.New(matchDevs, scanPeriod, *scanOnce, *printMode)
		} else {
			sw = xkb.New()
		}
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
		matchDevs, *scanOnce,
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
	for {
		select {
		case input := <-kbdLost:
			delete(keyboards, input)
		default:
			for devicePath, device := range getInputDevices() {
				if deviceMatch.MatchString(device.Name) {
					if _, ok := keyboards[devicePath]; !ok {
						if !quietMode {
							log.Printf("keyboard found at %s: %s", devicePath, device.Name)
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
			time.Sleep(scanPeriod)
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

	for {
		select {
		case msg := <-inbox:
			for _, ev := range msg.Events {
				if ev.Type != evdev.EV_KEY {
					continue
				}

				if printMode {
					pv := "released"
					if ev.Value == 1 {
						pv = "pressed"
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
