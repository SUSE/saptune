package system

import (
	"testing"
)

func TestSystemctlIsRunning(t *testing.T) {
	// check, if command is available
	if !CmdIsAvailable("/usr/bin/systemctl") {
		t.Skip("command '/usr/bin/systemctl' not available. Skip tests")
	}
	if !SystemctlIsRunning("dbus.service") {
		t.Log("'dbus.service' not running, skip test")
	}
	if !SystemctlIsRunning("tuned.service") {
		t.Log("'tuned.service' not running, skip test")
	}
}

func TestGetTunedProfile(t *testing.T) {
	actualVal := GetTunedProfile()
	if actualVal == "" {
		t.Log("seams there is no tuned profile, skip test")
	}
}
