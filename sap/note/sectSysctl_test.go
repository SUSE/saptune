package note

import (
	"github.com/SUSE/saptune/txtparser"
	"testing"
)

func TestOptSysctlVal(t *testing.T) {
	// remember the change in saptune 2.0 (SAP and Alliance decision)
	// use exactly the value from the config file. No calculation any more
	op := txtparser.Operator("=")
	val := OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300 200", "100 330 180")
	if val != "100	330	180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300", "100 330 180")
	if val != "" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "", "100 330 180")
	if val != "100 330 180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "PNA", "100 330 180")
	if val != "100 330 180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300 200", "")
	if val != "" {
		t.Error(val)
	}
	op = txtparser.Operator("<")
	val = OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "180")
	if val != "180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "120")
	if val != "120" {
		t.Error(val)
	}
	op = txtparser.Operator(">")
	val = OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "180")
	if val != "180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "120")
	if val != "120" {
		t.Error(val)
	}
}
