package system

import (
	"testing"
)

func TestIsUserRoot(t *testing.T) {
	if !IsUserRoot() {
		t.Log("the test requires root access")
	}
}

func TestGetOsName(t *testing.T) {
	actualVal := GetOsName()
	if actualVal != "SLES" {
		t.Logf("OS is '%s' and not 'SLES'\n", actualVal)
	}
	if actualVal == "" {
		t.Logf("empty value returned for the os Name")
	}
}

func TestGetOsVers(t *testing.T) {
	actualVal := GetOsVers()
	switch actualVal {
	case "12", "12-SP1", "12-SP2", "12-SP3", "12-SP4", "15", "15-SP1":
		t.Logf("expected OS version '%s' found\n", actualVal)
	default:
		t.Logf("unexpected OS version '%s'\n", actualVal)
	}
}
