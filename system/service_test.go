package system

import (
	"os"
	"testing"
)

func TestGetServiceName(t *testing.T) {
	value := GetServiceName("sysstat")
	if value != "sysstat.service" {
		t.Errorf("found service '%s' instead of 'sysstat.service'\n", value)
	}
	value = GetServiceName("sysstat.service")
	if value != "sysstat.service" {
		t.Errorf("found service '%s' instead of 'sysstat.service'\n", value)
	}
	value = GetServiceName("UnkownService")
	if value != "" {
		t.Errorf("found service '%s' instead of 'UnkownService'\n", value)
	}
}

func TestGetAvailServices(t *testing.T) {
	// test with missing command
	services = map[string]string{"": ""}
	cmdName := "/usr/bin/systemctl"
	savName := "/usr/bin/systemctl_SAVE"
	if err := os.Rename(cmdName, savName); err != nil {
		t.Error(err)
	}
	value := GetAvailServices()
	if len(value) != 0 {
		t.Error("found services")
	}
	service := GetServiceName("sysstat")
	if service != "" {
		t.Errorf("found service '%s' instead of 'UnkownService'\n", service)
	}
	if err := os.Rename(savName, cmdName); err != nil {
		t.Error(err)
	}
}

func TestCmpServiceStates(t *testing.T) {
	match := false
	current := "stop, disable"
	expected := "stop"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = ""
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "enable, stop"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "start, disable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "disable, stop"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "stop, start, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, stop, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, hugo, stop, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, hugo, stop, enable, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "sToP, hugo, start, disable, enable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "stop, hugo, start, enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "start, stop, enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "hugo"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
}
