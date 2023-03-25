package sway

import (
	"encoding/json"
	"os"
	"strings"
)

func CheckAvailability() bool {
	if os.Getenv("SWAYSOCK") != "" {
		return true
	}
	checkVars := []string{"DESKTOP_SESSION", "XDG_CURRENT_DESKTOP", "XDG_SESSION_DESKTOP"}
	for _, v := range checkVars {
		if strings.ToLower(os.Getenv(v)) == "sway" {
			return true
		}
	}
	return false
}

type SwayInput struct {
	Identifier           string `json:"identifier"`
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	XkbActiveLayoutIndex int    `json:"xkb_active_layout_index"`
	XkbActiveLayoutName  string `json:"xkb_active_layout_name"`
}

func GetInputDevices() ([]SwayInput, error) {
	out, err := swayexec("-t", "get_inputs")
	if err != nil {
		return nil, err
	}
	var inputs []SwayInput
	if err = json.Unmarshal(out, &inputs); err != nil {
		return nil, err
	}
	return inputs, nil
}
