package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// CheckCPUState csMap passend für Tests
func TestCheckCPUState(t *testing.T) {
	tstEqualMap := map[string]string{"cpu0": "state0:0 state1:0 state2:0 state3:0 state4:0", "cpu1": "state0:0 state1:0 state2:0 state3:0 state4:0", "cpu2": "state0:0 state1:0 state2:0 state3:0 state4:0", "cpu3": "state0:0 state1:0 state2:0 state3:0 state4:0"}
	tstDiffMap := map[string]string{"cpu0": "state0:0 state1:0 state2:0 state3:0 state4:0", "cpu1": "state0:0 state1:1 state2:0 state3:0 state4:0", "cpu2": "state0:0 state1:0 state2:1 state3:0 state4:0", "cpu3": "state0:0 state1:0 state2:0 state3:0 state4:1"}

	differ := checkCPUState(tstEqualMap)
	if differ {
		t.Error(differ)
	}
	differ = checkCPUState(tstDiffMap)
	if !differ {
		t.Error(differ)
	}
}

func TestSupportsPerfBias(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
	}

	if !supportsPerfBias() {
		t.Skip("System does not support Intel's performance bias setting. Skipping test")
	}
	cmdName := "/usr/bin/cpupower"
	cmdArgs := []string{"info", "-b"}

	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil || (err == nil && (strings.Contains(string(cmdOut), notSupportedX86) || strings.Contains(string(cmdOut), notSupportedIBM))) {
		t.Error(string(cmdOut))
	}
}

func TestGetPerfBias(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
	}
	value := GetPerfBias()
	if !supportsPerfBias() {
		if value != "all:none" {
			t.Error(value)
		}
	} else {
		if len(value) < 3 {
			t.Error(value)
		}
	}
}

func TestSetPerfBias(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
	}
	oldPerf := GetPerfBias()
	err := SetPerfBias("all:15")
	if err != nil {
		t.Error(err)
	}
	val := GetPerfBias()
	if val != "all:15" && val != "all:none" {
		t.Error(val)
	}
	if oldPerf != "" && oldPerf != "all:none" {
		// set test value back
		err := SetPerfBias(oldPerf)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestIsValidGovernor(t *testing.T) {
	_, err := os.ReadDir("/sys/devices/system/cpu/cpu0/cpufreq")
	if err != nil {
		t.Skip("directory '/sys/devices/system/cpu/cpu0/cpufreq' does not exist. System does not support scaling governor, skipping test")
	}
	gov, _ := GetSysString("devices/system/cpu/cpu0/cpufreq/scaling_governor")
	if !isValidGovernor("cpu0", gov) {
		t.Error(gov)
	}
	if isValidGovernor("not_avail", gov) {
		t.Error(gov)
	}
	if isValidGovernor("cpu0", "not_avail") {
		t.Errorf("governor 'not_avail' reported as supported, but shouldn't")
	}
}

func TestGetGovernor(t *testing.T) {
	_, err := os.ReadDir("/sys/devices/system/cpu/cpu0/cpufreq")
	if err != nil {
		t.Skip("directory '/sys/devices/system/cpu/cpu0/cpufreq' does not exist. System does not support scaling governor, skipping test")
	}
	gov, _ := GetSysString("devices/system/cpu/cpu0/cpufreq/scaling_governor")
	for k, v := range GetGovernor() {
		if k == "cpu0" && v != gov {
			t.Errorf("cpu0: expected '%s', actual '%s'\n", gov, v)
		}
		if k == "all" && v != gov {
			t.Errorf("all: expected '%s', actual '%s'\n", gov, v)
		}
	}
}

func TestSetGovernor(t *testing.T) {
	oldGov := GetGovernor()
	gov := "performance"
	err := SetGovernor("all:performance")
	if err != nil {
		t.Error(err)
	}
	for k, v := range GetGovernor() {
		if k == "all" && (v != gov && v != "none") {
			t.Errorf("all: expected '%s', actual '%s'\n", gov, v)
		}
	}
	err = SetGovernor("cpu0:performance")
	if err != nil {
		t.Error(err)
	}
	for k, v := range GetGovernor() {
		if k == "cpu0" && (v != gov && v != "none") {
			t.Errorf("cpu0: expected '%s', actual '%s'\n", gov, v)
		}
	}
	// set test value back
	val := ""
	for k, v := range oldGov {
		val = val + fmt.Sprintf("%s:%s ", k, v)
	}
	err = SetGovernor(val)
	if err != nil {
		t.Error(err)
	}
	err = SetGovernor("cpu0:performance")
	if err != nil {
		t.Error(err)
	}
}

func TestGetdmaLatency(t *testing.T) {
	value := GetdmaLatency()
	if value == "" {
		t.Log("/dev/cpu_dma_latency is not supported")
	} else {
		t.Log(value)
	}
}

func TestGetFLInfo(t *testing.T) {
	value, _, _ := GetFLInfo()
	if value == "all:none" {
		t.Log("system does not support force_latency settings")
	} else {
		t.Log(value)
	}
}

func TestSetForceLatency(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
	}
	oldLat, _, _ := GetFLInfo()
	err := SetForceLatency("all:none", "cpu1:state0:0 cpu1:state1:0", false)
	if err != nil {
		t.Error(err)
	}
	err = SetForceLatency("70", "cpu1:state0:0 cpu1:state1:0", false)
	if err != nil {
		t.Error(err)
	}
	err = SetForceLatency("70", "cpu1:state0:0 cpu1:state1:0", false)
	t.Log(err)
	err = SetForceLatency("70", "cpu1:state0:0 cpu1:state1:0", true)
	t.Log(err)

	if oldLat != "" {
		// set test value back
		err := SetForceLatency(oldLat, "", false)
		if err != nil {
			t.Error(err)
		}
	}
}

// test with missing cpupower command
func TestMissingCmd(t *testing.T) {
	cmdName := "/usr/bin/cpupower"
	savName := "/usr/bin/cpupower_SAVE"
	if err := os.Rename(cmdName, savName); err != nil {
		t.Error(err)
	}
	value := GetPerfBias()
	if value != "all:none" {
		t.Error(value)
	}
	if err := SetPerfBias("all:15"); err != nil {
		t.Error(err)
	}
	if supportsPerfBias() {
		t.Errorf("reports supported, but shouldn't")
	}
	if err := SetGovernor("all:performance"); err != nil {
		t.Error(err)
	}
	if err := os.Rename(savName, cmdName); err != nil {
		t.Error(err)
	}
}

func TestCPUErrorCases(t *testing.T) {
	oldCpupowerCmd := cpupowerCmd
	defer func() { cpupowerCmd = oldCpupowerCmd }()
	cpupowerCmd = "/usr/bin/false"
	val := GetPerfBias()
	if val != "all:none" {
		t.Error(val)
	}
	if err := SetPerfBias("all:15"); err != nil {
		t.Errorf("should return 'nil' and not '%v'\n", err)
	}
	if isValidGovernor("cpu0", "performance") {
		if err := SetGovernor("all:performance"); err == nil {
			t.Error("should return an error and not 'nil'")
		}
	} else {
		if err := SetGovernor("all:performance"); err != nil {
			t.Errorf("should return 'nil' and not '%v'\n", err)
		}
	}
	if supportsPerfBias() {
		t.Error("reports supported, but shouldn't")
	}
	cpupowerCmd = oldCpupowerCmd

	oldCPUDir := cpuDir
	defer func() { cpuDir = oldCPUDir }()
	cpuDir = "/unknownDir"
	gval := GetGovernor()
	if len(gval) != 1 {
		t.Errorf("should return only one entry, but returns: %+v", gval)
	}
	for k, v := range gval {
		if k != "all" && v != "none" {
			t.Errorf("expected 'all:none', actual '%s:%s'\n", k, v)
		}
	}
	if supportsForceLatencySettings("70") {
		if err := SetForceLatency("70", "cpu1:state0:0 cpu1:state1:0", false); err == nil {
			t.Error("should return an error and not 'nil'")
		}
	}
	cpuDir = oldCPUDir
}
