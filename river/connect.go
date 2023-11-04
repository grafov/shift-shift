package river

import (
	"bytes"
	"os/exec"
)

func riverctl(command ...string) ([]byte, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("riverctl", command...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	err := cmd.Run()
	return stdout.Bytes(), err
}
