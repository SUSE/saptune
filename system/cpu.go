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

// constant definition
const (
	notSupportedX86 = "System does not support Intel's performance bias setting"
	notSupportedIBM = "Subcommand not supported on POWER."
	cpuDirSys       = "devices/system/cpu"
)

var efiVarsDir = "/sys/firmware/efi/efivars"
var cpuDir = "/sys/devices/system/cpu"
var cpupowerCmd = "/usr/bin/cpupower"
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
		case line == notSupportedX86 || line == notSupportedIBM:
			return "all:none"
		case isPBCpu.MatchString(line):
			str = str + fmt.Sprintf("cpu%d", k/2)
		case isPBias.MatchString(line):
			pb := strings.Split(line, ":")
			if len(pb) < 2 {
				continue
			}
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
	if !canSetPerfBias() {
		return nil
	}

	cpu := ""
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if len(fields) < 2 {
			continue
		}
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
		} else {
			cpu = fields[0]
		}
		out, err := exec.Command(cpupowerCmd, "-c", cpu, "set", "-b", fields[1]).CombinedOutput()
		if err != nil {
			WarningLog("failed to invoke external command 'cpupower -c %s set -b %s': %v, output: %s", cpu, fields[1], err, out)
			return err
		}
	}
	return nil
}

// canSetPerfBias checks, if Perf Bias can be set
func canSetPerfBias() bool {
	setPerf := true
	if GetCSP() == "azure" {
		WarningLog("Setting Perf Bias is not supported on '%s'\n", CSPAzureLong)
		setPerf = false
	} else if SecureBootEnabled() {
		WarningLog("Cannot set Perf Bias when SecureBoot is enabled, skipping")
		setPerf = false
	} else if !supportsPerfBias() {
		WarningLog("Perf Bias settings not supported by the system")
		setPerf = false
	}
	return setPerf
}

// SecureBootEnabled checks, if the system is in lock-down mode
func SecureBootEnabled() bool {
	var isSecBootFileName = regexp.MustCompile(`^SecureBoot-\w[\w-]+`)
	if _, err := os.Stat(efiVarsDir); os.IsNotExist(err) {
		InfoLog("no EFI directory '%+s' found, assuming legacy boot", efiVarsDir)
		return false
	}
	secureBootFile := ""
	_, efiFiles := ListDir(efiVarsDir, "the available efi variables")
	for _, eFile := range efiFiles {
		if isSecBootFileName.MatchString(eFile) {
			// work with the first file matching 'SecureBoot-*'
			secureBootFile = path.Join(efiVarsDir, eFile)
			break
		}
	}
	if secureBootFile == "" {
		InfoLog("no EFI SecureBoot file (SecureBoot-*) found in '%s', assuming legacy boot", efiVarsDir)
		return false
	}

	content, err := ioutil.ReadFile(secureBootFile)
	if err != nil {
		InfoLog("failed to read EFI SecureBoot file '%s': %v", secureBootFile, err)
		return false
	}
	lastElement := content[len(content)-1]
	if lastElement == 1 {
		DebugLog("secure boot enabled - '%v'", content)
		return true
	}
	DebugLog("secure boot disabled - '%v'", content)
	return false
}

// supportsPerfBias check, if the system will support CPU performance settings
func supportsPerfBias() bool {
	cmdName := cpupowerCmd
	cmdArgs := []string{"info", "-b"}

	if !CmdIsAvailable(cmdName) {
		WarningLog("command '%s' not found", cmdName)
		return false
	}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil || (err == nil && (strings.Contains(string(cmdOut), notSupportedX86) || strings.Contains(string(cmdOut), notSupportedIBM))) {
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
				tmpfile := path.Join(cpuDir, entry.Name(), "cpufreq", "scaling_governor")
				InfoLog("Unable to identify the current scaling governor for CPU '%s', missing file '%s'. Check your intel_pstate.", entry.Name(), tmpfile)
				gov = ""
			} else {
				// GetSysString needs cpuDirSys as path - without /sys
				gov, _ = GetSysString(path.Join(cpuDirSys, entry.Name(), "cpufreq", "scaling_governor"))
			}
			if gov == "" || gov == "NA" || gov == "PNA" {
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
	if !canSetGov(value, info) {
		return nil
	}

	cpu := ""
	tst := ""
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if len(fields) < 2 {
			continue
		}
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
			tst = cpu
		} else {
			cpu = fields[0]
			tst = "cpu0"
		}
		if !isValidGovernor(tst, fields[1]) {
			WarningLog("'%s' is not a valid governor, skipping.", fields[1])
			continue
		}
		out, err := exec.Command(cpupowerCmd, "-c", cpu, "frequency-set", "-g", fields[1]).CombinedOutput()
		if err != nil {
			WarningLog("failed to invoke external command 'cpupower -c %s frequency-set -g %s': %v, output: %s", cpu, fields[1], err, out)
			return err
		}
	}
	return nil
}

// canSetGov checks, if the governor can be set
func canSetGov(value, info string) bool {
	setGov := true
	if GetCSP() == "azure" {
		WarningLog("Setting governor is not supported on '%s'\n", CSPAzureLong)
		setGov = false
	} else if value == "all:none" || info == "notSupportedX86" || info == "notSupportedIBM" {
		WarningLog("governor settings not supported by the system")
		setGov = false
	} else if !CmdIsAvailable(cpupowerCmd) {
		WarningLog("command '%s' not found", cpupowerCmd)
		setGov = false
	}
	return setGov
}

// isValidGovernor check, if the system will support CPU frequency settings
func isValidGovernor(cpu, gov string) bool {
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
	cpuStateDiffer = checkCPUState(cpuStateMap)

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

	if !canSetForceLatency(value, info) {
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
			cpudirCont, errns := ioutil.ReadDir(path.Join(cpuDir, entry.Name(), "cpuidle"))
			if errns != nil {
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
							if len(FLFields) > 2 {
								if FLFields[0] == entry.Name() && FLFields[1] == centry.Name() {
									oldState = FLFields[2]
								}
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
						if lat > flval {
							// set new latency states
							err = SetSysString(path.Join(cpuDirSys, entry.Name(), "cpuidle", centry.Name(), "disable"), "1")
						}
						if lat <= flval && oldState == "1" {
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

// canSetForceLatency checks, if Force Latency can be set
func canSetForceLatency(value, info string) bool {
	setLatency := true
	if GetCSP() == "azure" {
		WarningLog("latency settings are not supported on '%s'\n", CSPAzureLong)
		setLatency = false
	}
	if value == "all:none" || info == "notSupportedX86" || info == "notSupportedIBM" {
		WarningLog("latency settings not supported by the system")
		setLatency = false
	}
	return setLatency
}

// checkCPUState checks, if all cpus have the same state settings
// returns true, if the cpu states differ
func checkCPUState(csMap map[string]string) bool {
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
