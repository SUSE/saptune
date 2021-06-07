package note

import (
	"testing"
)

func TestGetGrubVal(t *testing.T) {
	val := GetGrubVal("grub:processor.max_cstate")
	if val == "NA" {
		t.Log("'processor.max_cstate' not found in kernel cmdline")
	}
	val = GetGrubVal("grub:UNKNOWN")
	if val != "NA" {
		t.Error(val)
	}
}

func TestOptGrubVal(t *testing.T) {
	val := OptGrubVal("grub:processor.max_cstate", "NO_OPT")
	if val != "NO_OPT" {
		t.Error(val)
	}
}

func TestSetGrubVal(t *testing.T) {
	val := SetGrubVal("NO_OPT")
	if val != nil {
		t.Error(val)
	}
}
