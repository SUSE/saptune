package note

import (
	"testing"
)

func TestGetVMVal(t *testing.T) {
	val, _ := GetVMVal("THP")
	if val != "always" && val != "madvise" && val != "never" {
		t.Errorf("wrong value '%+v' for THP.\n", val)
	}
	val, _ = GetVMVal("KSM")
	if val != "1" && val != "0" {
		t.Errorf("wrong value '%+v' for KSM.\n", val)
	}
}

func TestOptVMVal(t *testing.T) {
	val := OptVMVal("THP", "always")
	if val != "always" {
		t.Error(val)
	}
	val = OptVMVal("THP", "unknown")
	if val != "never" {
		t.Error(val)
	}
	val = OptVMVal("KSM", "1")
	if val != "1" {
		t.Error(val)
	}
	val = OptVMVal("KSM", "2")
	if val != "0" {
		t.Error(val)
	}
	val = OptVMVal("UNKOWN_PARAMETER", "unknown")
	if val != "unknown" {
		t.Error(val)
	}
}

func TestSetVMVal(t *testing.T) {
	newval := ""
	oldval, _ := GetVMVal("THP")
	if oldval == "never" {
		newval = "always"
	} else {
		newval = "never"
	}
	err := SetVMVal("THP", newval)
	if err != nil {
		t.Error(err)
	}
	val, _ := GetVMVal("THP")
	if val != newval {
		t.Error(val)
	}
	// set test value back
	err = SetVMVal("THP", oldval)
	if err != nil {
		t.Error(err)
	}

	oldval, _ = GetVMVal("KSM")
	if oldval == "0" {
		newval = "1"
	} else {
		newval = "0"
	}
	err = SetVMVal("KSM", newval)
	if err != nil {
		t.Error(err)
	}
	val, _ = GetVMVal("KSM")
	if val != newval {
		t.Error(val)
	}
	// set test value back
	err = SetVMVal("KSM", oldval)
	if err != nil {
		t.Error(err)
	}
}
