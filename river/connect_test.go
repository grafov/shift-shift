package river

import "testing"

func TestCommand(t *testing.T) {
	_, err := riverctl("list-inputs")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
