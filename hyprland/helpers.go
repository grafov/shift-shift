package hyprland

import (
	"fmt"
	"os"
	"strings"

	"github.com/thiagokokada/hyprland-go"
	"github.com/thiagokokada/hyprland-go/helpers"
)

type keyboard struct {
	Address      string `json:"address"`
	Name         string `json:"name"`
	Rules        string `json:"rules"`
	Model        string `json:"model"`
	Layout       string `json:"layout"`
	Variant      string `json:"variant"`
	Options      string `json:"options"`
	ActiveKeymap string `json:"active_keymap"`
	Main         bool   `json:"main"`
}

func PrintDevices() ([]string, error) {
	s, err := helpers.GetSocket(helpers.RequestSocket)
	if err != nil {
		return nil, err
	}
	inputs, err := getDevices(hyprland.NewClient(s))
	if err != nil {
		return nil, err
	}
	output := make([]string, len(inputs))
	for i, inp := range inputs {
		if !inp.Main {
			continue
		}
		output[i] = fmt.Sprintf("model:%s name:%s layout:%s options:%s keymap:%s\n",
			inp.Model, inp.Name, inp.Layout, inp.Options, inp.ActiveKeymap)
	}
	return output, nil
}

func getDevices(h *hyprland.RequestClient) ([]keyboard, error) {
	devs, err := h.Devices()
	if err != nil {
		return nil, err
	}
	kbds := make([]keyboard, len(devs.Keyboards))
	for i, v := range devs.Keyboards {
		kbds[i] = keyboard{
			Address:      v.Address,
			Name:         v.Name,
			Rules:        v.Rules,
			Model:        v.Model,
			Layout:       v.Layout,
			Variant:      v.Variant,
			Options:      v.Options,
			ActiveKeymap: v.ActiveKeymap,
			Main:         v.Main,
		}
	}
	return kbds, nil
}

func CheckAvailability() bool {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		return true
	}
	checkVars := []string{"DESKTOP_SESSION", "XDG_CURRENT_DESKTOP", "XDG_SESSION_DESKTOP"}
	for _, v := range checkVars {
		if strings.ToLower(os.Getenv(v)) == "hyprland" {
			return true
		}
	}
	return false
}
