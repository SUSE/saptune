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

func TestWriteSysctl(t *testing.T) {
	oldval, err := GetSysctlInt("vm.max_map_count")
	if err != nil {
		t.Fatal(err)
	}
	if err := SetSysctlInt("vm.max_map_count", 65630); err != nil {
		t.Fatal(err)
	}
	intval, err := GetSysctlInt("vm.max_map_count")
	if intval != 65630 {
		t.Fatal(intval)
	}
	if err := SetSysctlUint64("vm.max_map_count", 65635); err != nil {
		t.Fatal(err)
	}
	uintval, err := GetSysctlUint64("vm.max_map_count")
	if uintval != 65635 {
		t.Fatal(uintval)
	}
	if err := SetSysctlString("vm.max_map_count", "65640"); err != nil {
		t.Fatal(err)
	}
	sval, err := GetSysctlString("vm.max_map_count")
	if sval != "65640" {
		t.Fatal(sval)
	}

	oldfield, err := GetSysctlUint64Field("net.ipv4.ip_local_port_range", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetSysctlUint64Field("net.ipv4.ip_local_port_range", 0, 31768); err != nil {
		t.Fatal(err)
	}
	uintval, err = GetSysctlUint64Field("net.ipv4.ip_local_port_range", 0)
	if uintval != 31768 {
		t.Fatal(uintval)
	}

	if err := SetSysctlString("UnknownKey", "100"); err != nil {
		t.Fatal(err)
	}
	if err := SetSysctlUint64Field("UnknownKey", 1, 100); err == nil {
		t.Fatal(err)
	}
	// set test value back
	if err := SetSysctlInt("vm.max_map_count", oldval); err != nil {
		t.Fatal(err)
	}
	if err := SetSysctlUint64Field("net.ipv4.ip_local_port_range", 0, oldfield); err != nil {
		t.Fatal(err)
	}
}

func TestIsPagecacheAvailable(t *testing.T) {
	if IsPagecacheAvailable() {
		t.Log("pagecache setting available")
	} else {
		t.Log("pagecache setting NOT available")
	}
}
