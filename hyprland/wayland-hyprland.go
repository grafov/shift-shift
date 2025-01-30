package hyprland

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/thiagokokada/hyprland-go"
	"github.com/thiagokokada/hyprland-go/helpers"
)

type Hyprland struct {
	client *hyprland.RequestClient
	re     *regexp.Regexp
	sleep  time.Duration
	once   bool
	debug  bool

	m         sync.RWMutex
	keyboards []string
}

func New(re *regexp.Regexp, scanPeriod time.Duration, scanOnce bool, debug bool) (*Hyprland, error) {
	s, err := helpers.GetSocket(helpers.RequestSocket)
	if err != nil {
		return nil, err
	}
	return &Hyprland{client: hyprland.NewClient(s), re: re, sleep: scanPeriod, once: scanOnce, debug: debug}, nil
}

func (h *Hyprland) Init() error {
	go h.matchKeyboards(h.debug)
	return nil
}

func (h *Hyprland) Switch(id int) {
	h.m.RLock()
	for _, kbd := range h.keyboards {
		if h.debug {
			log.Printf("switch hyprland kbd \"%s\" to group %d", kbd, id-1)
		}
		resp, err := h.client.SwitchXkbLayout(kbd, strconv.Itoa(id-1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "response: %v, error: %s", resp, err)
		}
	}
	h.m.RUnlock()
}

func (h *Hyprland) Name() string {
	return "Hyprland"
}

func (h *Hyprland) Close() {}

func (h *Hyprland) matchKeyboards(debug bool) {
	for {
		inputs, err := getDevices(h.client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't get input devices from Hyprland: %s", err)
		}
		h.m.Lock()
		h.keyboards = nil
		for _, in := range inputs {
			if h.re.MatchString(in.Name) && in.Main {
				if debug {
					log.Printf("Hyprland keyboard matched %s", in.Name)
				}

				h.keyboards = append(h.keyboards, in.Name)
			}
		}
		h.m.Unlock()
		if h.once {
			return
		}
		time.Sleep(h.sleep)
	}
}
