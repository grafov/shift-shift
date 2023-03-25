package sway

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const swayKbdType = "keyboard"

type Sway struct {
	re    *regexp.Regexp
	sleep time.Duration
	once  bool
	debug bool

	m         sync.Mutex
	keyboards []string
}

func New(re *regexp.Regexp, scanPeriod time.Duration, scanOnce bool, debug bool) *Sway {
	return &Sway{re: re, sleep: scanPeriod, once: scanOnce, debug: debug}
}

func (s *Sway) Init() error {
	go s.matchOnlyKeyboards()
	return nil
}

func (*Sway) Name() string {
	return "sway"
}

func (s *Sway) Switch(idx int) {
	var totalcmd strings.Builder
	for _, n := range s.keyboards {
		if s.debug {
			log.Printf("switch sway kbd \"%s\" to group %d", n, idx-1)
		}
		totalcmd.WriteString(fmt.Sprintf("input %s xkb_switch_layout %d;", n, idx-1))
	}
	if _, err := swayexec(totalcmd.String()); err != nil {
		fmt.Fprintf(os.Stderr, "sway error on switching: %s\n", err)
	}
}

func (*Sway) Close() {}

func (s *Sway) matchOnlyKeyboards() {
	for {
		inputs, err := GetInputDevices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't get input devices from Sway: %s", err)
		}
		s.m.Lock()
		s.keyboards = nil
		for _, i := range inputs {
			if s.re.MatchString(i.Name) && i.Type == swayKbdType {
				s.keyboards = append(s.keyboards, i.Identifier)
			}
		}
		s.m.Unlock()
		if s.once {
			return
		}
		time.Sleep(s.sleep)
	}
}
