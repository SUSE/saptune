package system

import "testing"

func TestReadSysctl(t *testing.T) {
	if value, err := GetSysctlInt("vm.max_map_count"); err != nil {
		t.Fatal(value)
	}
	if value, err := GetSysctlUint64("vm.max_map_count"); err != nil {
		t.Fatal(value)
	}
	if value, _ := GetSysctlString("vm.max_map_count"); len(value) < 2 { // indeed testing string length
		t.Fatal(value)
	}

	if value, err := GetSysctlInt("does not exist"); err == nil {
		t.Fatal(value)
	}
	if value, err := GetSysctlUint64("does not exist"); err == nil {
		t.Fatal(value)
	}
	if value, err := GetSysctlString("does not exist"); err == nil {
		t.Fatal(value)
	}
}
