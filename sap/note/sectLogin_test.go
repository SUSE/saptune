package note

import (
	"github.com/SUSE/saptune/system"
	"os"
	"testing"
)

func TestGetLoginVal(t *testing.T) {
	val, err := GetLoginVal("Unknown")
	if val != "" || err != nil {
		t.Error(val)
	}

	val, err = GetLoginVal("UserTasksMax")
	if _, errno := os.Stat("/etc/systemd/logind.conf.d/saptune-UserTasksMax.conf"); errno != nil {
		if !os.IsNotExist(errno) {
			if val != "" || err == nil {
				t.Error(val)
			}
		} else {
			if val != "NA" || err != nil {
				t.Error(val)
			}
		}
	}
}

func TestOptLoginVal(t *testing.T) {
	val := OptLoginVal("unknown")
	if val != "unknown" {
		t.Error(val)
	}
	val = OptLoginVal("infinity")
	if val != "infinity" {
		t.Error(val)
	}
	val = OptLoginVal("")
	if val != "" {
		t.Error(val)
	}
}

func TestSetLoginVal(t *testing.T) {
	utmFile := "/etc/systemd/logind.conf.d/saptune-UserTasksMax.conf"
	val := "18446744073709"

	err := SetLoginVal("UserTasksMax", val, false)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(utmFile); err != nil {
		t.Error(err)
	}
	if !system.CheckForPattern(utmFile, val) {
		t.Errorf("wrong value in file '%s'\n", utmFile)
	}
	val = "infinity"
	err = SetLoginVal("UserTasksMax", val, false)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(utmFile); err != nil {
		t.Error(err)
	}
	if !system.CheckForPattern(utmFile, val) {
		t.Errorf("wrong value in file '%s'\n", utmFile)
	}
	val = "10813"
	err = SetLoginVal("UserTasksMax", val, true)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(utmFile); err == nil {
		os.Remove(utmFile)
		t.Errorf("file '%s' still exists\n", utmFile)
	}
}
