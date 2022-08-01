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
