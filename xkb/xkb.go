package xkb

// #cgo LDFLAGS: -lX11
// #include <stdlib.h>
// #include <stdio.h>
// #include <err.h>
// #include <X11/Xlib.h>
// #include <X11/XKBlib.h>
import "C"

import (
	"fmt"
	"log"
)

type Xkb struct {
	display *C.Display
}

func New() *Xkb {
	return &Xkb{}
}

func (*Xkb) Name() string {
	return "xkb"
}

func (x *Xkb) Init() error {
	var err error
	x.display, err = openDisplay()
	if err != nil {
		return fmt.Errorf("unable to open X display: %w\n", err)
	}
	return nil
}

func (x *Xkb) Switch(idx int) {
	var group uint
	switch idx {
	case 1:
		group = C.XkbGroup1Index
	case 2:
		group = C.XkbGroup2Index
	case 3:
		group = C.XkbGroup3Index
	case 4:
		group = C.XkbGroup4Index
	default:
		group = C.XkbGroup1Index
	}
	result := C.XkbLockGroup(x.display, C.XkbUseCoreKbd, C.uint(group))
	if result != 1 {
		log.Println("unable to send lock group request to X11")
		return
	}
	// immideately output events buffer
	C.XFlush(x.display)
}

func (x *Xkb) Close() {
	C.XCloseDisplay(x.display)
}

func openDisplay() (*C.Display, error) {
	var xkbEventType, xkbError, xkbReason C.int
	var majorVers, minorVers C.int

	majorVers = C.XkbMajorVersion
	minorVers = C.XkbMinorVersion

	display := C.XkbOpenDisplay(
		nil, &xkbEventType, &xkbError, &majorVers, &minorVers, &xkbReason,
	)
	if display == nil {
		switch xkbReason {
		case C.XkbOD_BadServerVersion:
		case C.XkbOD_BadLibraryVersion:
			return nil, fmt.Errorf("incompatible versions of client and server XKB libraries")
		case C.XkbOD_ConnectionRefused:
			return nil, fmt.Errorf("connection to X server refused")
		case C.XkbOD_NonXkbServer:
			return nil, fmt.Errorf("XKB extension is not present")
		default:
			return nil, fmt.Errorf("unknown error from XkbOpenDisplay: %d", xkbReason)
		}
	}

	return display, nil
}
