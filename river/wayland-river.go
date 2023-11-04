package river

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

type River struct {
	re    *regexp.Regexp
	sleep time.Duration
	once  bool
	debug bool
}

func New(re *regexp.Regexp, scanPeriod time.Duration, scanOnce bool, debug bool) *River {
	return &River{re: re, sleep: scanPeriod, once: scanOnce, debug: debug}
}

func (s *River) Init() error {
	return nil
}

func (*River) Name() string {
	return "river"
}

// Switch in very hacky way yet. Hardcoded for us/ru layouts! Till I
// understand how to properly switch layouts through Wayland with xkb.
func (s *River) Switch(idx int) {
	var lt string
	switch idx {
	case 1:
		lt = "us"
	case 2:
		lt = "ru"
	default:
		lt = "us"
	}
	if s.debug {
		log.Printf("switch river to group %s", lt)
	}
	out, err := riverctl("keyboard-layout", lt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "river error on switching: %s\n", err)
	}
	if s.debug {
		fmt.Fprintf(os.Stderr, string(out))
	}
}

func (*River) Close() {}
