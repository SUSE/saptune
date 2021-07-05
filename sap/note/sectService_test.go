package note

import (
	"github.com/SUSE/saptune/system"
	"strings"
	"testing"
)

var services = system.GetAvailServices()

func TestGetServiceName(t *testing.T) {
	val := system.GetServiceName("uuidd.socket")
	if val != "uuidd.socket" && val != "" {
		t.Error(val)
	}
	val = system.GetServiceName("sysstat")
	if val != "sysstat.service" && val != "" {
		t.Error(val)
	}
	val = system.GetServiceName("sysstat.service")
	if val != "sysstat.service" && val != "" {
		t.Error(val)
	}
	val = system.GetServiceName("UnkownService")
	if val != "" {
		t.Error(val)
	}
}

func TestGetServiceVal(t *testing.T) {
	wrong := false
	state := ""
	val := GetServiceVal("UnkownService")
	if val != "NA" {
		t.Error(val)
	}
	val = GetServiceVal("uuidd.socket")
	for _, st := range strings.Split(val, ",") {
		state = strings.TrimSpace(st)
		if state != "start" && state != "stop" && state != "NA" && state != "enable" && state != "disable" {
			wrong = true
		}
	}
	if wrong {
		t.Error(val)
	}
}

func TestOptServiceVal(t *testing.T) {
	val := OptServiceVal("UnkownService", "start")
	if val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("uuidd.socket", "start")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("uuidd.socket", "stop")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("uuidd.socket", "unknown")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("sysstat", "start")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("sysstat.service", "stop")
	if val != "stop" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("sysstat", "unknown")
	if val != "" && val != "NA" {
		t.Error(val)
	}
	wrong := false
	state := ""
	val = OptServiceVal("sysstat", "stop, start, unknown, disable, enable")
	for _, st := range strings.Split(val, ",") {
		state = strings.TrimSpace(st)
		if state != "stop" && state != "disable" && state != "NA" {
			wrong = true
		}
	}
	if wrong {
		t.Error(val)
	}
	wrong = false
	val = OptServiceVal("uuidd.socket", "enable")
	for _, st := range strings.Split(val, ",") {
		state = strings.TrimSpace(st)
		if state != "start" && state != "enable" && state != "NA" {
			wrong = true
		}
	}
	if wrong {
		t.Error(val)
	}
}

func TestSetServiceVal(t *testing.T) {
	val := SetServiceVal("UnkownService", "start")
	if val != nil {
		t.Error(val)
	}
	_ = system.SystemctlDisable("sysstat.service")
	val = SetServiceVal("sysstat.service", "enable")
	if val != nil {
		t.Error(val)
	}
	val = SetServiceVal("sysstat.service", "disable")
	if val != nil {
		t.Error(val)
	}
}
