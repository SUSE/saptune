package system

import "testing"

func TestReadSysctl(t *testing.T) {
	if value, err := GetSysctlInt("vm.max_map_count"); err != nil {
		t.Error(value)
	}
	if value, err := GetSysctlUint64("vm.max_map_count"); err != nil {
		t.Error(value)
	}
	if value, _ := GetSysctlString("vm.max_map_count"); len(value) < 2 { // indeed testing string length
		t.Error(value)
	}
	if value, err := GetSysctlUint64Field("net.ipv4.ip_local_port_range", 0); err != nil {
		t.Error(value, err)
	} else {
		t.Log(value)
	}

	if value, err := GetSysctlInt("does not exist"); err == nil {
		t.Error(value)
	}
	if value, err := GetSysctlUint64("does not exist"); err == nil {
		t.Error(value)
	}
	if value, err := GetSysctlString("does not exist"); err == nil {
		t.Error(value)
	}

	if value, err := GetSysctlUint64Field("does not exist", 0); err == nil {
		t.Error(value, err)
	}
}

func TestWriteSysctl(t *testing.T) {
	oldval, err := GetSysctlInt("vm.max_map_count")
	if err != nil {
		t.Error(err)
	}
	if err := SetSysctlInt("vm.max_map_count", 65630); err != nil {
		t.Error(err)
	}
	intval, err := GetSysctlInt("vm.max_map_count")
	if err != nil {
		t.Error(err)
	}
	if intval != 65630 {
		t.Error(intval)
	}
	if err := SetSysctlUint64("vm.max_map_count", 65635); err != nil {
		t.Error(err)
	}
	uintval, err := GetSysctlUint64("vm.max_map_count")
	if err != nil {
		t.Error(err)
	}
	if uintval != 65635 {
		t.Error(uintval)
	}
	if err := SetSysctlString("vm.max_map_count", "65640"); err != nil {
		t.Error(err)
	}
	sval, err := GetSysctlString("vm.max_map_count")
	if err != nil {
		t.Error(err)
	}
	if sval != "65640" {
		t.Error(sval)
	}

	oldfield, err := GetSysctlUint64Field("net.ipv4.ip_local_port_range", 0)
	if err != nil {
		t.Error(err)
	}
	if err := SetSysctlUint64Field("net.ipv4.ip_local_port_range", 0, 31768); err != nil {
		t.Error(err)
	}
	uintval, err = GetSysctlUint64Field("net.ipv4.ip_local_port_range", 0)
	if err != nil {
		t.Error(err)
	}
	if uintval != 31768 {
		t.Error(uintval)
	}

	if err := SetSysctlString("vm.dirty_bytes", "100"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if err := SetSysctlString("net.ipv4.ip_local_port_range", "PNA"); err != nil {
		t.Error(err)
	}
	if err := SetSysctlString("UnknownKey", "100"); err != nil {
		t.Error(err)
	}
	// net.ipv4.ip_local_port_range has only 2 fields
	if err := SetSysctlUint64Field("net.ipv4.ip_local_port_range", 3, 4711); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if err := SetSysctlUint64Field("UnknownKey", 1, 100); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	// set test value back
	if err := SetSysctlInt("vm.max_map_count", oldval); err != nil {
		t.Error(err)
	}
	if err := SetSysctlUint64Field("net.ipv4.ip_local_port_range", 0, oldfield); err != nil {
		t.Error(err)
	}
}

func TestIsPagecacheAvailable(t *testing.T) {
	if IsPagecacheAvailable() {
		t.Log("pagecache setting available")
	} else {
		t.Log("pagecache setting NOT available")
	}
}

func TestGlobalSysctls(t *testing.T) {
	//var sysctlParms = sysctlDefined{}
	expTxt := "sysctl config file /etc/sysctl.d/saptune_test.conf(1), /etc/sysctl.d/saptune_test2.conf(1)"
	expTxt2 := "sysctl config file /etc/sysctl.d/saptune_test2.conf(1), /etc/sysctl.d/saptune_test.conf(1)"
	CollectGlobalSysctls()
	info := ChkForSysctlDoubles("vm.nr_hugepages")
	if info != "" {
		t.Errorf("got '%s' instead of expected empty string\n", info)
	}
	info = ChkForSysctlDoubles("vm.pagecache_limit_ignore_dirty")
	if info == "" {
		t.Error("got empty string instead of expected text")
	}
	// as the order of the sysctl files are not predictive and not important
	// for the code use 2 text pattern for the comparison
	if info != expTxt && info != expTxt2 {
		t.Errorf("got '%s' instead of expected text '%s'\n", info, expTxt)
	}
}
