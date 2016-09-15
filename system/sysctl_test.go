package system

import "testing"

func TestReadSysctl(t *testing.T) {
	if value := GetSysctlInt("vm.max_map_count", 0); value < 10 {
		t.Fatal(value)
	}
	if value := GetSysctlUint64("vm.max_map_count", 0); value < 10 {
		t.Fatal(value)
	}
	if value := GetSysctlString("vm.max_map_count", "0"); len(value) < 2 { // indeed testing string length
		t.Fatal(value)
	}

	if value := GetSysctlInt("does not exist", 123); value != 123 {
		t.Fatal(value)
	}
	if value := GetSysctlUint64("does not exist", 123); value != 123 {
		t.Fatal(value)
	}
	if value := GetSysctlString("does not exist", "default"); value != "default" {
		t.Fatal(value)
	}
}
