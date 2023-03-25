package sway

import "testing"

func TestCommand(t *testing.T) {
	_, err := swayexec("-t", "get_inputs")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
