package note

import (
	"testing"
)

func TestGetCPUVal(t *testing.T) {
	val, _, _ := GetCPUVal("force_latency")
	if val != "all:none" {
		t.Logf("force_latency supported: '%s'\n", val)
	}
	val, _, _ = GetCPUVal("energy_perf_bias")
	if val != "all:none" {
		t.Logf("energy_perf_bias supported: '%s'\n", val)
	}
	val, _, _ = GetCPUVal("governor")
	if val != "all:none" && val != "" {
		t.Logf("governor supported: '%s'\n", val)
	}
}

func TestOptCPUVal(t *testing.T) {
	val := OptCPUVal("force_latency", "1000", "70")
	if val != "70" {
		t.Error(val)
	}

	val = OptCPUVal("energy_perf_bias", "all:15", "performance")
	if val != "all:0" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "cpu0:15 cpu1:6 cpu2:0", "performance")
	if val != "cpu0:0 cpu1:0 cpu2:0" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:15", "normal")
	if val != "all:6" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:15", "powersave")
	if val != "all:15" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:15", "unknown")
	if val != "all:0" {
		t.Error(val)
	}

	/* future feature
	val = OptCPUVal("energy_perf_bias", "cpu0:6 cpu1:6 cpu2:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:0 cpu1:6 cpu2:15" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:performance cpu1:normal cpu2:powersave" {
		t.Error(val)
	}
	*/

	val = OptCPUVal("governor", "all:powersave", "performance")
	if val != "all:performance" {
		t.Error(val)
	}
	val = OptCPUVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "performance")
	if val != "cpu0:performance cpu1:performance cpu2:performance" {
		t.Error(val)
	}
	/* future feature
	val = OptCPUVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Error(val)
	}
	*/
}

//SetCPUVal
