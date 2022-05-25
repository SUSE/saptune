package note

import (
	"github.com/SUSE/saptune/txtparser"
	"testing"
)

func TestGetSysVal(t *testing.T) {
	val, inf := GetSysVal("sys:kernel/mm/ksm/run")
	if val != "0" {
		t.Logf("expected '0', got '%s'\n", val)
	}
	if inf != "" {
		t.Errorf("expected '', got '%s'\n", inf)
	}
	val, inf = GetSysVal("kernel.mm.transparent_hugepage.enabled")
	if val != "always" {
		t.Logf("expected 'always', got '%s'\n", val)
	}
	if inf != "" {
		t.Errorf("expected '', got '%s'\n", inf)
	}
	val, inf = GetSysVal("kernel.mm.angi")
	if val != "PNA" {
		t.Errorf("expected 'PNA', got '%s'\n", val)
	}
	if inf != "" {
		t.Errorf("expected '', got '%s'\n", inf)
	}
}

func TestOptSysVal(t *testing.T) {
	op := txtparser.Operator("=")
	key := "sys:TestParam"
	val := OptSysVal(op, key, "120", "100")
	if val != "100" {
		t.Error(val)
	}
	key = "TestParam"
	val = OptSysVal(op, key, "120 300 200", "100 330 180")
	if val != "100	330	180" {
		t.Error(val)
	}
}
