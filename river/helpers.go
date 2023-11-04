package river

import (
	"os"
	"strings"
)

func CheckAvailability() bool {
	checkVars := []string{"DESKTOP_SESSION", "XDG_CURRENT_DESKTOP", "XDG_SESSION_DESKTOP"}
	for _, v := range checkVars {
		if strings.ToLower(os.Getenv(v)) == "river" {
			return true
		}
	}
	return false
}
