package sway

import (
	"encoding/json"
	"fmt"
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

type device struct {
	Identifier           string `json:"identifier"`
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	XkbActiveLayoutIndex int    `json:"xkb_active_layout_index"`
	XkbActiveLayoutName  string `json:"xkb_active_layout_name"`
}

func getDevices() ([]device, error) {
	out, err := swayexec("-t", "get_inputs")
	if err != nil {
		return nil, err
	}
	var inputs []device
	if err = json.Unmarshal(out, &inputs); err != nil {
		return nil, err
	}
	return inputs, nil
}

func PrintDevices() ([]string, error) {
	inputs, err := getDevices()
	if err != nil {
		return nil, err
	}
	output := make([]string, len(inputs))
	for i, inp := range inputs {
		output[i] = fmt.Sprintf("%s type:%s name:%s layout:%s group:%d\n",
			inp.Identifier, inp.Type, inp.Name, inp.XkbActiveLayoutName, inp.XkbActiveLayoutIndex)
	}
	return output, err
}
