package system

import (
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
)


func TestSupportsPerfBias(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
	}

	if !SupportsPerfBias() {
		t.Skip("System does not support Intel's performance bias setting. Skipping test")
	}
	cmdName := "/usr/bin/cpupower"
	cmdArgs := []string{"info", "-b"}

	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil || (err == nil && strings.Contains(string(cmdOut), notSupported)) {
		t.Fatal(string(cmdOut))
	}
}

func TestGetPerfBias(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
	}
	value := GetPerfBias()
	if !SupportsPerfBias() {
		if value != "all:none" {
			t.Fatal(value)
		}
	} else {
		if len(value) < 3 {
			t.Fatal(value)
		}
	}
}

func TestIsValidGovernor(t *testing.T) {
	_, err := ioutil.ReadDir("/sys/devices/system/cpu/cpu0/cpufreq")
	if err != nil {
		t.Skip("directory '/sys/devices/system/cpu/cpu0/cpufreq' does not exist. System does not support scaling governor, skipping test")
	}
	gov, _ := GetSysString("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
	if !IsValidGovernor("cpu0", gov) {
		t.Fatal(gov)
	}
}

func TestGetGovernor(t *testing.T) {
	_, err := ioutil.ReadDir("/sys/devices/system/cpu/cpu0/cpufreq")
	if err != nil {
		t.Skip("directory '/sys/devices/system/cpu/cpu0/cpufreq' does not exist. System does not support scaling governor, skipping test")
	}
	gov, _ := GetSysString("devices/system/cpu/cpu0/cpufreq/scaling_governor")
	for k, v := range GetGovernor() {
		if k == "cpu0" && v != gov {
			t.Fatalf("cpu0: expected '%s', actual '%s'\n", gov, v)
		}
		if k == "all" && v != gov {
			t.Fatalf("all: expected '%s', actual '%s'\n", gov, v)
		}
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

func TestGetForceLatency(t *testing.T) {
	value := GetForceLatency()
	if value == "all:none" {
		t.Log("system does not support force_latency settings")
	} else {
		t.Log(value)
	}
}
