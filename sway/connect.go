package sway

import (
	"bytes"
	"os/exec"
)

const swaymsg = "swaymsg"

func swayexec(command ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(swaymsg, command...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), err
}
