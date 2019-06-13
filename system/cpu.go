package system

// wrapper to cpupower command

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

//constant definition
const (
	notSupported = "System does not support Intel's performance bias setting"
	cpuDir       = "/sys/devices/system/cpu"
	cpuDirSys    = "devices/system/cpu"
	cpupowerCmd  = "/usr/bin/cpupower"
)

var isCPU = regexp.MustCompile(`^cpu\d+$`)
var isState = regexp.MustCompile(`^state\d+$`)

// GetPerfBias retrieve CPU performance configuration from the system
func GetPerfBias() string {
	isPBCpu := regexp.MustCompile(`analyzing CPU \d+`)
	isPBias := regexp.MustCompile(`perf-bias: \d+`)
	setAll := true
	str := ""
	oldpb := "99"
	cmdName := cpupowerCmd
	cmdArgs := []string{"-c", "all", "info", "-b"}

	if !CmdIsAvailable(cmdName) {
		WarningLog("command '%s' not found", cmdName)
		return "all:none"
	}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		WarningLog("There was an error running external command 'cpupower -c all info -b': %v, output: %s", err, cmdOut)
		return "all:none"
	}

	for k, line := range strings.Split(strings.TrimSpace(string(cmdOut)), "\n") {
		switch {
		case line == notSupported:
			//WarningLog(notSupported)
			return "all:none"
		case isPBCpu.MatchString(line):
			str = str + fmt.Sprintf("cpu%d", k/2)
		case isPBias.MatchString(line):
			pb := strings.Split(line, ":")
			if oldpb == "99" {
				oldpb = strings.TrimSpace(pb[1])
			}
			if oldpb != strings.TrimSpace(pb[1]) {
				setAll = false
			}
			str = str + fmt.Sprintf(":%s ", strings.TrimSpace(pb[1]))
		}
	}
	if setAll {
		str = "all:" + oldpb
	}
	return strings.TrimSpace(str)
}

// SetPerfBias set CPU performance configuration to the system using 'cpupower' command
func SetPerfBias(value string) error {
	//cmd := exec.Command("cpupower", "-c", "all", "set", "-b", value)
	cpu := ""
	if !SupportsPerfBias() {
		WarningLog(notSupported)
		return nil
	}
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
		} else {
			cpu = fields[0]
		}
		cmd := exec.Command(cpupowerCmd, "-c", cpu, "set", "-b", fields[1])
		out, err := cmd.CombinedOutput()
		if err != nil {
			WarningLog("failed to invoke external command 'cpupower -c %s set -b %s': %v, output: %s", cpu, fields[1], err, out)
			return err
		}
	}
	return nil
}

// SupportsPerfBias check, if the system will support CPU performance settings
func SupportsPerfBias() bool {
	cmdName := cpupowerCmd
	cmdArgs := []string{"info", "-b"}

	if !CmdIsAvailable(cmdName) {
		WarningLog("command '%s' not found", cmdName)
		return false
	}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil || (err == nil && strings.Contains(string(cmdOut), notSupported)) {
		// does not support perf bias
		return false
	}
	return true
}

// GetGovernor retrieve performance configuration regarding to cpu frequency
// from the system
func GetGovernor() map[string]string {
	setAll := true
	oldgov := "99"
	gov := ""
	gGov := make(map[string]string)

	dirCont, err := ioutil.ReadDir(cpuDir)
	if err != nil {
		return gGov
	}
	for _, entry := range dirCont {
		if isCPU.MatchString(entry.Name()) {
			if _, err = os.Stat(path.Join(cpuDir, entry.Name(), "cpufreq", "scaling_governor")); os.IsNotExist(err) {
				// os.Stat needs cpuDir as path - including /sys
				gov = ""
			} else {
				// GetSysString needs cpuDirSys as path - without /sys
				gov, _ = GetSysString(path.Join(cpuDirSys, entry.Name(), "cpufreq", "scaling_governor"))
			}
			if gov == "" {
				gov = "none"
			}
			if oldgov == "99" {
				// starting point
				oldgov = gov
			}
			if oldgov != gov {
				setAll = false
			}
			gGov[entry.Name()] = gov
		}
	}
	if setAll {
		gGov = make(map[string]string)
		gGov["all"] = oldgov
	}
	return gGov
}

// SetGovernor set performance configuration regarding to cpu frequency
// to the system using 'cpupower' command
func SetGovernor(value, info string) error {
	//cmd := exec.Command("cpupower", "-c", "all", "frequency-set", "-g", value)
	cpu := ""
	tst := ""
	cmdName := cpupowerCmd

	if value == "all:none" || info == "notSupported" {
		WarningLog("governor settings not supported by the system")
		return nil
	}
	if !CmdIsAvailable(cmdName) {
		WarningLog("command '%s' not found", cmdName)
		return nil
	}
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
			tst = cpu
		} else {
			cpu = fields[0]
			tst = "cpu0"
		}
		if !IsValidGovernor(tst, fields[1]) {
			WarningLog("'%s' is not a valid governor, skipping.", fields[1])
			continue
		}
		cmd := exec.Command(cpupowerCmd, "-c", cpu, "frequency-set", "-g", fields[1])
		out, err := cmd.CombinedOutput()
		if err != nil {
			WarningLog("failed to invoke external command 'cpupower -c %s frequency-set -g %s': %v, output: %s", cpu, fields[1], err, out)
			return err
		}
	}
	return nil
}

// IsValidGovernor check, if the system will support CPU frequency settings
func IsValidGovernor(cpu, gov string) bool {
	val, err := ioutil.ReadFile(path.Join(cpuDir, cpu, "/cpufreq/scaling_available_governors"))
	if err == nil && strings.Contains(string(val), gov) {
		return true
	}
	return false
}

// GetFLInfo retrieve CPU latency configuration from the system and returns
// the current latency,
// the latency states of all CPUs to save latency states for 'revert',
// if cpu states differ
// return lat, savedStates, cpuStateDiffer
func GetFLInfo() (string, string, bool) {
	lat := 0
	maxlat := 0
	supported := false
	savedStates := ""
	stateDisabled := false
	cpuStateDiffer := false
	cpuStateMap := make(map[string]string)

	// read /sys/devices/system/cpu
	dirCont, err := ioutil.ReadDir(cpuDir)
	if runtime.GOARCH != "ppc64le" && err == nil {
		// latency settings are only relevant for Intel-based systems
		for _, entry := range dirCont {
			// cpu0 ... cpuXY
			if isCPU.MatchString(entry.Name()) {
				// read /sys/devices/system/cpu/cpu*/cpuidle
				cpudirCont, err := ioutil.ReadDir(path.Join(cpuDir, entry.Name(), "cpuidle"))
				if err != nil {
					// idle settings not supported for entry.Name()
					continue
				}
				supported = true
				for _, centry := range cpudirCont {
					// state0 ... stateXY
					if isState.MatchString(centry.Name()) {
						// read /sys/devices/system/cpu/cpu*/cpuidle/state*/disable
						state, _ := GetSysString(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "disable"))
						// save latency states for 'revert'
						// savedStates = "cpu1:state0:0 cpu1:state1:0"
						savedStates = savedStates + " " + entry.Name() + ":" + centry.Name() + ":" + state
						cpuStateMap[entry.Name()] = cpuStateMap[entry.Name()] + " " + state
						// read /sys/devices/system/cpu/cpu*/cpuidle/state*/latency
						lattmp, _ := GetSysInt(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "latency"))
						if lattmp > maxlat {
							maxlat = lattmp
						}
						if state == "1" {
							stateDisabled = true
						} else {
							lat = lattmp
						}
					}
				}
			}
		}
	}
	// check, if all cpus have the same state settings
	cpuStateDiffer = CheckCPUState(cpuStateMap)

	if !stateDisabled {
		// start value for force latency, if no states are disabled
		lat = maxlat
	}

	rval := strconv.Itoa(lat)
	if !supported {
		savedStates = "all:none"
		rval = "all:none"
	}
	return rval, savedStates, cpuStateDiffer
}

// SetForceLatency set CPU latency configuration to the system
func SetForceLatency(value, savedStates, info string, revert bool) error {
	oldState := ""

	if value == "all:none" || info == "notSupported" {
		WarningLog("latency settings not supported by the system")
		return nil
	}

	flval, _ := strconv.Atoi(value) // decimal value for force latency

	dirCont, err := ioutil.ReadDir(cpuDir)
	if err != nil {
		WarningLog("latency settings not supported by the system")
		return err
	}
	for _, entry := range dirCont {
		// cpu0 ... cpuXY
		if isCPU.MatchString(entry.Name()) {
			cpudirCont, err := ioutil.ReadDir(path.Join(cpuDir, entry.Name(), "cpuidle"))
			if err != nil {
				WarningLog("idle settings not supported for '%s'", entry.Name())
				continue
			}
			for _, centry := range cpudirCont {
				// state0 ... stateXY
				if isState.MatchString(centry.Name()) {
					// read /sys/devices/system/cpu/cpu*/cpuidle/state*/latency
					lat, _ := GetSysInt(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "latency"))
					// write /sys/devices/system/cpu/cpu*/cpuidle/state*/disable
					if revert {
						// revert
						for _, ole := range strings.Fields(savedStates) {
							FLFields := strings.Split(ole, ":")
							if FLFields[0] == entry.Name() && FLFields[1] == centry.Name() {
								oldState = FLFields[2]
							}
						}
						if oldState != "" {
							err = SetSysString(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "disable"), oldState)
							// clear latency value for next cpu/state cycle
							oldState = ""
						}
					} else {
						// apply
						oldState, _ = GetSysString(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "disable"))
						// save old latency states for 'revert'
						if lat >= flval {
							// set new latency states
							err = SetSysString(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "disable"), "1")
						}
						if lat < flval && oldState == "1" {
							// reset previous set latency state
							err = SetSysString(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "disable"), "0")
						}
					}
				}
			}
		}
	}

	return err
}

// CheckCPUState checks, if all cpus have the same state settings
// returns true, if the cpu states differ
func CheckCPUState(csMap map[string]string) bool {
	ret := false
	oldcpuState := ""
	for _, cpuState := range csMap {
		if oldcpuState == "" {
			oldcpuState = cpuState
		}
		if oldcpuState != cpuState {
			ret = true
			break
		}
	}
	return ret
}

// GetdmaLatency retrieve DMA latency configuration from the system
func GetdmaLatency() string {
	latency := make([]byte, 4)
	dmaLatency, err := os.OpenFile("/dev/cpu_dma_latency", os.O_RDONLY, 0600)
	if err != nil {
		WarningLog("GetForceLatency: failed to open cpu_dma_latency - %v", err)
	}
	_, err = dmaLatency.Read(latency)
	if err != nil {
		WarningLog("GetForceLatency: reading from '/dev/cpu_dma_latency' failed: %v", err)
	}
	// Close the file handle after the latency value is no longer maintained
	defer dmaLatency.Close()

	ret := fmt.Sprintf("%v", binary.LittleEndian.Uint32(latency))
	return ret
}
